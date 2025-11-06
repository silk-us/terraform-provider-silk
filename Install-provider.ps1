param(
    [string] $version = '1.2.5'
)

$releaseURI = 'https://github.com/silk-us/terraform-provider-silk/releases/download/v.' + $version + '/terraform-provider-silk_' + $version + '_windows_amd64.exe'

$providerPath = $env:appdata + '\terraform.d\plugins\localdomain\provider\silk\' + $version + '\windows_amd64\'
New-Item -Path $providerPath -ItemType  Directory -Force
Invoke-RestMethod -Uri $releaseURI -OutFile ($providerPath + "terraform-provider-silk.exe")
