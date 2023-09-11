using System;
using System.EnterpriseServices;
using System.Runtime.InteropServices;
using System.IO;
using System.Net;
using System.Diagnostics;
using System.Security.Cryptography.X509Certificates;
using System.Net.Http;
using System.Security.Authentication;
using System.Web.Script.Serialization;

namespace rancher.gmsa
{

    [Guid("6ECDA518-2010-4437-8BC3-46E752B7B172")]
    [InterfaceType(ComInterfaceType.InterfaceIsIUnknown)]
    [ComImport]
    public interface ICcgDomainAuthCredentials
    {
        void GetPasswordCredentials(
            [MarshalAs(UnmanagedType.LPWStr), In] string pluginInput,
            [MarshalAs(UnmanagedType.LPWStr)] out string domainName,
            [MarshalAs(UnmanagedType.LPWStr)] out string username,
            [MarshalAs(UnmanagedType.LPWStr)] out string password);
    }

    [Guid("e4781092-f116-4b79-b55e-28eb6a224e26")]
    [ProgId("CcgCredProvider")]
    public class CcgCredProvider : ServicedComponent, ICcgDomainAuthCredentials
    {

        public void GetPasswordCredentials(
            [MarshalAs(UnmanagedType.LPWStr), In] string pluginInput,
            [MarshalAs(UnmanagedType.LPWStr)] out string domainName,
            [MarshalAs(UnmanagedType.LPWStr)] out string username,
            [MarshalAs(UnmanagedType.LPWStr)] out string password)
        {
            ServicePointManager.Expect100Continue = true;

            PluginInput decodedInput = DecodeInput(pluginInput);
            SetupLogger(decodedInput);

            bool certsAvailable = false;
            var crt = "/var/lib/rancher/gmsa/" + decodedInput.ActiveDirectory + "/ssl/client/tls.pfx";
            if (File.Exists(crt)) {
                certsAvailable = true;
            }

            try
            {
                ResponseBody response;
                if (certsAvailable)
                {
                    response = GetCredential(decodedInput);
                }
                else
                {
                    response = GetUnverifiedCredential(decodedInput);
                }
                domainName = response.DomainName;
                username = response.UserName;
                password = response.Password;
            }
            catch (Exception e)
            {
                // log the exception ourself
                // so we know we can find it
                LogError(e.ToString());
                // throw it again so ccg can catch it
                // and print its own error logs
                throw;
            }
        }

        // Queries the account provider API with mTLS certificates
        public ResponseBody GetCredential(PluginInput pluginInput)
        {

            // we don't have a PKI setup to distribute CRL, so disable the check globally
            ServicePointManager.CheckCertificateRevocationList = false;

            X509Certificate2 clientCertificate = new X509Certificate2(File.ReadAllBytes(pluginInput.GetCertFile()), (string)null, X509KeyStorageFlags.MachineKeySet);

            try
            {
                HttpClient httpClient = new HttpClient(new HttpClientHandler
                {
                    ClientCertificateOptions = ClientCertificateOption.Manual,
                    SslProtocols = SslProtocols.Tls12,
                    ClientCertificates = { clientCertificate },
                    CheckCertificateRevocationList = false,
                });

                var responseBody = MakeRequest(httpClient, pluginInput);

                // creating x509Certificate2 objects writes a few files to disk,
                // make sure we clean them up now that we are done with them
                clientCertificate.Reset();
                clientCertificate.Dispose();

                return new JavaScriptSerializer().Deserialize<ResponseBody>(responseBody);
            }
            catch (Exception ex)
            {
                LogError("Http Client Hit An Exception: \n " + ex.ToString());
            }

            clientCertificate.Reset();
            clientCertificate.Dispose();
            return null;
        }

        // Queries the account provider API without mTLS certificates
        public ResponseBody GetUnverifiedCredential(PluginInput pluginInput)
        {
            var secretUri = "https://localhost:" + pluginInput.Port + "/provider";
            try
            {
                HttpClient httpClient = new HttpClient(new HttpClientHandler{});

                var responseBody = MakeRequest(httpClient, pluginInput);

                return new JavaScriptSerializer().Deserialize<ResponseBody>(responseBody);
            }
            catch (Exception ex)
            {
                LogError("Http Client Hit An Exception: \n " + ex.ToString());
            }
            return null;
        }

        public string MakeRequest(HttpClient client, PluginInput pluginInput)
        {
            var secretUri = "https://localhost:" + pluginInput.Port + "/provider";
            LogInfo("Preparing to make request: Using secret: " + pluginInput.SecretName + " from namespace: " + pluginInput.ActiveDirectory + " and port: " + pluginInput.Port + " results in uri: " + secretUri);

            var httpRequestMessage = new HttpRequestMessage(HttpMethod.Get, secretUri);
            httpRequestMessage.Headers.Add("object", pluginInput.SecretName);
            var response = client.SendAsync(httpRequestMessage).Result;
            var responseBody = response.Content.ReadAsStringAsync().Result;
            LogInfo("Got response with content of: " + responseBody);
            return responseBody;
        }

        // logger is our Event Logger. We log to the Application source, not a custom source
        // this allows us to circumvent privileged operations required to setup a new source
        private EventLog logger;
        private bool writeLogs;
        public CcgCredProvider()
        {
            logger = new EventLog("Application");
            logger.Source = "Application";
        }

        private void LogInfo(string log)
        {
            if (!writeLogs)
            {
                return;
            }
            logger.WriteEntry(log, EventLogEntryType.Information, 101, 1);
        }

        private void LogWarn(string log)
        {
            if (!writeLogs)
            {
                return;
            }
            logger.WriteEntry(log, EventLogEntryType.Warning, 201, 1);
        }

        private void LogError(string log)
        {
            logger.WriteEntry(log, EventLogEntryType.Error, 301, 1);
        }

        private void SetupLogger(PluginInput decodedInput)
        {
            writeLogs = File.Exists(decodedInput.GetDebugFile());
        }

        public PluginInput DecodeInput(string pluginInput)
        {
            return new PluginInput(pluginInput);
        }

        public class ResponseBody
        {
            public string UserName { get; set; }
            public string Password { get; set; }
            public string DomainName { get; set; }
        }

        public class PluginInput
        {
            public PluginInput(string pluginInput)
            {
                var parts = pluginInput.Split(':');
                if (parts.Length != 2)
                {
                    throw new Exception("Invalid Plugin Input Format");
                }
                this.ActiveDirectory = parts[0];
                this.SecretName = parts[1];
                this.Port = GetPort();
            }

            public string ActiveDirectory { get; set; }
            public string SecretName { get; set; }
            public string Port { get; set; }

            public string GetPort()
            {
                try
                {
                    return File.ReadAllText(GetPortFile());
                }
                catch (Exception e)
                {
                    throw new Exception("Failed to open port file located at " + GetPortFile() + ": " + e.ToString());
                }
            }

            public string GetPortFile()
            {
                return "/var/lib/rancher/gmsa/" + this.ActiveDirectory + "/port.txt";
            }

            public string GetCertFile()
            {
                // we use a pfx so we can bundle the cert and private key together in a single file
                return "/var/lib/rancher/gmsa/" + this.ActiveDirectory + "/ssl/client/tls.pfx";
            }

            public string GetDebugFile()
            {
                return "/var/lib/rancher/gmsa/" + this.ActiveDirectory + "/enable-logs.txt";
            }
        }
    }
}
