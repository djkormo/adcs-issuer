## How to install ADCS on VM using Hyper-V

While installing both Active Directory Domain Services (AD DS) and Active Directory Certificate Services (AD CS) on the same server is perfectly fine for a lab or testing environment, Microsoft strongly advises against this in production. It creates a single point of failure and makes disaster recovery significantly more complex.


Starting from a fresh installation of Windows Server Standard, you must follow a strict order: install the domain controller, promote it, reboot, log in with domain credentials, and then install the certificate authority.

Using mainly Powershell scripts

### Install AD DS Binaries

```powershell
Install-WindowsFeature -Name AD-Domain-Services -IncludeManagementTools
```


Verification: The progress bar will complete, and the Success column in the output will display True.


### Promote to Domain Controller

Promote the server to a Domain Controller to create your new Active Directory forest. Replace "djkormo.github.com" with your desired domain name.

```powershell
Install-ADDSForest -DomainName "djkormo.github.com" -InstallDns -Force
```


(Note: You will be prompted in the console to enter and confirm a SafeModeAdministratorPassword. This is required for Directory Services Restore Mode, not your daily admin password).
Verification: Once the configuration is applied, the server will automatically reboot.


### Log in as Domain Administrator

After the reboot, do not log in with the local administrator account. You must log in using the new Domain Admin credentials (e.g., djkormo.github.com\Administrator). 

The Enterprise CA installation will fail if you do not have Enterprise Admin privileges, which are only granted to this domain account.


### Install AD CS Binaries


Once logged into the desktop as the domain administrator, open a new elevated PowerShell window and install the Certificate Services role.


```powershell
Install-WindowsFeature -Name AD-Certificate -IncludeManagementTools
```

Verification: The Success column in the output will display True.

### Configure the Enterprise Root CA

Configure the server as an Enterprise Root Certificate Authority. Because this server is now a Domain Controller, it can seamlessly integrate with Active Directory.

```powershell
Install-AdcsCertificationAuthority `
  -CAType EnterpriseRootCa `
  -CryptoProviderName "RSA#Microsoft Software Key Storage Provider" `
  -KeyLength 2048 `
  -HashAlgorithmName SHA256 `
  -ValidityPeriod Years `
  -ValidityPeriodUnits 10 `
  -Force
```


Verification: Run Get-Service certsvc. The status should output as Running, confirming your Certificate Authority is active and online.


To add the Web Enrollment portal to your existing Certificate Authority, you need to install the IIS-related AD CS binaries and then configure the service to bind with your local CA.

Open an elevated PowerShell console (Run as Administrator) and follow these steps:


### Install the Web Enrollment Binaries


First, add the Windows Feature that downloads the necessary IIS and AD CS web components.

```powershell
Install-WindowsFeature -Name ADCS-Web-Enrollment
```

Verification: Wait for the progress bar to finish. The output table should display True under the Success column and No under Restart Needed.

### Configure the Web Enrollment Role

Next, run the deployment cmdlet to configure the virtual directory in IIS and link it to your active Certificate Authority. The -Force parameter suppresses the confirmation prompt.


```powershell
Install-AdcsWebEnrollment -Force
```

Verification: The command will execute quickly and return to the prompt without red error text, indicating the IIS application was successfully created.


### Restart the Web Server

To ensure the new application pool and virtual directory are fully loaded, restart the Internet Information Services (IIS) service.

```powershell
iisreset
```
Verification: The console should output "Internet services successfully restarted."


### Verify the Portal

Open a web browser on your server or a domain-joined machine and navigate to the local portal:

http://localhost/certsrv
(Note: The certsrv directory is case-sensitive in some IIS configurations, so keep it lowercase).Verification: 
You will be prompted to authenticate. Enter your domain administrator credentials. 
Once logged in, you should see the "Welcome" page for Microsoft Active Directory Certificate Services.

Note on HTTPS: By default, Web Enrollment installs using HTTP. To fully utilize the portal from other computers (and to avoid browser security blocks when requesting certain types of certificates), you will eventually need to bind an SSL certificate to your "Default Web Site" in the IIS Manager.


### How to add templates

An empty Certificate Templates console typically happens in new environments when the Active Directory Certificate Services (AD CS) installation finishes before the new Active Directory Domain Services (AD DS) partition has fully initialized, meaning the default templates never get written to the directory.

While there isn't a native PowerShell cmdlet specifically for this action, you can call the built-in Windows certificate utility directly from your elevated PowerShell console to force the synchronization.

#### Force Template Creation

Ensure you are running PowerShell as an Administrator and are logged in with your Domain Admin credentials. Run the following command:

```powershell
certutil -installdefaulttemplates
```

Verification: The console will output a long list of template names being successfully added to the directory (e.g., Added User, Added WebServer, Added Computer) and will finish with a command completed successfully message.

#### Restart the Templates Console


If you still have the empty Certificate Templates console open, close it. Launch a fresh instance by typing the following in your PowerShell window or the Run dialog:

```powershell
certtmpl.msc
```

Verification: The console will open and should now be populated with 30+ default certificate templates ready for use.





### Creating new user in AD

```powershell
# 1. Convert plain text password to a secure string
$Password = ConvertTo-SecureString "Str0ngP@ssw0rd!" -AsPlainText -Force

# 2. Create the new domain user
New-ADUser -Name "CertRequester" -SamAccountName "certrequester" -AccountPassword $Password -Enabled $true

# 3. Create a new security group specifically for template access
New-ADGroup -Name "WebTemplateUsers" -GroupCategory Security -GroupScope Global

# 4. Add the new user to the security group
Add-ADGroupMember -Identity "WebTemplateUsers" -Members "certrequester"
```

Verification: Run 
```powershell
Get-ADGroupMember -Identity "WebTemplateUsers"
```
The output should list your new user (certrequester), confirming they are in the group.


#### Testing permissions
```powershell
Start-Process powershell.exe -Credential "corp\certrequester"
```

Verification: A Windows Security dialog will pop up asking for the password (Str0ngP@ssw0rd!). Once entered, a new PowerShell window will open with a blue background.


#### Attempt the enrollment

Get-Certificate.In the new PowerShell window (which is now running as certrequester), run the following command to request a certificate using your custom template. We will use a dummy Subject Name just for this test, and place it in the user's personal certificate store.

```powershell
Get-Certificate -Template "CorpWebServer" -CertStoreLocation "Cert:\CurrentUser\My" -SubjectName "CN=TestPermission"
```

Verification:

    Success: If permissions are correct, the command will immediately output the Thumbprint and Subject of a newly generated certificate.

    Failure: If the user does not have permission, you will receive a red error stating Access is denied (0x80070005) or The requested certificate template is not supported by this CA.


### Remove the test certificate:Cleanup (Optional).


If the test was successful, you now have a useless test certificate sitting in the user's personal store. You can safely delete it by running this in the same test window:

```powershell
Get-ChildItem -Path "Cert:\CurrentUser\My" | Where-Object Subject -match "CN=TestPermission" | Remove-Item
```
