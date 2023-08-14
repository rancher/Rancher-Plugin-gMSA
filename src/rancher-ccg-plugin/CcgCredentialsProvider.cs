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

    // TODO; env vars. How can we deploy this DLL in 'dev mode' so that we can more easily debug and
    // assess issues? Primary benefit would be enable / disable event logs. We should ensure event logs are
    // disabled when deployed to a real environment.

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
        // logger is our Event Logger. We log to the Application source, not a custom source
        // this allows us to circumvent privileged operations required to setup a new source
        private EventLog logger;
        public CcgCredProvider()
        {
            logger = new EventLog("Application");
            logger.Source = "Application";
        }

        private void LogInfo(string log)
        {
            logger.WriteEntry(log, EventLogEntryType.Information, 101, 1);
        }

        private void LogWarn(string log)
        {
            logger.WriteEntry(log, EventLogEntryType.Warning, 201, 1);
        }

        private void LogError(string log)
        {
            logger.WriteEntry(log, EventLogEntryType.Error, 301, 1);
        }

        public void GetPasswordCredentials(
            [MarshalAs(UnmanagedType.LPWStr), In] string pluginInput,
            [MarshalAs(UnmanagedType.LPWStr)] out string domainName,
            [MarshalAs(UnmanagedType.LPWStr)] out string username,
            [MarshalAs(UnmanagedType.LPWStr)] out string password)
        {
            ServicePointManager.Expect100Continue = true;
            try
            {
                var response = GetCredential(DecodeInput(pluginInput));
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

        public ResponseBody GetCredential(PluginInput pluginInput)
        {
            var secretUri = "https://localhost:" + pluginInput.Port + "/provider";
            // we don't have a PKI setup to distribute CRL, so disable the check globally
            ServicePointManager.CheckCertificateRevocationList = false;

            // we use a pfx so we can bundle the cert and private key together in a single file
            var crt = "/var/lib/rancher/gmsa/" + pluginInput.ActiveDirectory + "/ssl/client/tls.pfx";
            X509Certificate2 clientCertificate = new X509Certificate2(File.ReadAllBytes(crt), (string)null, X509KeyStorageFlags.MachineKeySet);
            LogInfo("Preparing to make request: Using secret: " + pluginInput.SecretName + "from namespace: " + pluginInput.ActiveDirectory + " and port: " + pluginInput.Port + " results in uri: " + secretUri);

            try
            {
                HttpClient httpClient = new HttpClient(new HttpClientHandler
                {
                    ClientCertificateOptions = ClientCertificateOption.Manual,
                    SslProtocols = SslProtocols.Tls12,
                    ClientCertificates = { clientCertificate },
                    CheckCertificateRevocationList = false,
                });

                var httpRequestMessage = new HttpRequestMessage(HttpMethod.Get, secretUri);
                httpRequestMessage.Headers.Add("object", pluginInput.SecretName);

                var response = httpClient.SendAsync(httpRequestMessage).Result;
                var x = response.Content.ReadAsStringAsync().Result;
                LogInfo("Got response, " + response.Content.ToString() + ", and content of: " + x);

                // creating x509Certificate2 objects writes a few files to disk,
                // make sure we clean them up now that we are done with them
                clientCertificate.Reset();
                clientCertificate.Dispose();

                return new JavaScriptSerializer().Deserialize<ResponseBody>(x);
            }
            catch (Exception ex)
            {
                LogError("Http Client Hit An Exception: \n " + ex.ToString());
            }
            clientCertificate.Reset();
            clientCertificate.Dispose();
            return null;
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
                string subDirFile = "/var/lib/rancher/gmsa/" + this.ActiveDirectory + "/port.txt";
                try
                {
                    return File.ReadAllText(subDirFile);
                }
                catch (Exception e)
                {
                    throw new Exception("Failed to open port file located at " + subDirFile + ": " + e.ToString());
                }
            }
        }
    }
}
