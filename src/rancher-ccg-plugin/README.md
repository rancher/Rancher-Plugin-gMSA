# C# gMSA CCG Plugin

This directory contains the source code of the Rancher gMSA CCG Plugin DLL, which written in C# and uses DotNet Framework v4.8.  

## Making Changes
In order to properly make changes to this component you will need to create a development environment in a Windows environment. This is due to the fact that the plugin is written as a .NET Framework application, due to its inherit use of Windows specific APIs and standards (COM+).  

### Setting up the IDE  

To get the best development experience you should use Visual Studio Community Edition. Visual Studio Community will help install and manage the various SDKs and Windows development tools needed to properly work with the plugin source code. While it is possible to use VSCode, you will need to install all required dependencies manually and lose useful features such as intellisense. 

When setting up Visual Studio, ensure the following libraries are also installed 
+ .NET Framework 4.8 SDK
+ C# and Visual Basic
+ C# and Visual Basic Roslyn compilers
+ .NET SDK 
+ All Windows 10 SDKs (versions 10.0.20348.0, 10.0.19041.0, and 10.0.18362.0)
+ .NET Compiler Platform SDK

To open the component in Visual Studio, simply double-click on the **Visual Studio Solution File** (rancher-gmsa.sln). 

### Building the Plugin
In order to build the Plugin as a DLL artifact, simply run `dotnet build`. By default, the artifact will be placed in `./bin/Debug/`. If building in Visual Studio, you can click on the `Build` tab in the upper section of the window, then select `Build rancher-gmsa`. 

### Best Practices And Requirements
This component is written in C#, and as such proper naming conventions and coding standards should be followed. A good list of C# coding standards can be found in [the Microsoft documentation](https://learn.microsoft.com/en-us/dotnet/csharp/fundamentals/coding-style/coding-conventions).

In addition to the general language best practices, more specific requirements and limitations exist for the plugin

1. The plugin should _only_ use the .NET Framework standard library. Using NuGet packages or other libraries will result in multiple DLL's being produced during the build process. This project aims to provide a single artifact to be used by CCG, so additional libraries should not be used unless they can be embedded within the final DLL. 
2. The GUID's specified within the application must **never** change. 
3. The `ICcgDomainAuthCredentials` interface should **never** change. You should not modify the existing method signature nor add new methods to the interface unless you have a good reason to do so. Improper changes may result in an improper COM object and thus break the DLL. 
4. The resulting DLL needs to be strong signed before it can be used in a cluster. The `Properties/AssemblyInfo.cs` already specifies the strong key file to be used. 
   1. In the future this project will move towards [delayed signing](https://learn.microsoft.com/en-us/dotnet/standard/assembly/delay-sign), at which point this process will change.
