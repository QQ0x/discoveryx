set -e

name=innovation2007
app_name=Innovation2007.app

cd bin

ditto -c -k --keepParent ${app_name} ${name}_darwin_amd64.zip # Use ditto instead of zip to keep metadata
zip ${name}_linux_386.zip ${name}_linux_386
zip ${name}_linux_amd64.zip ${name}_linux_amd64
zip ${name}_windows_386.zip ${name}_windows_386.exe steam_api.dll
zip ${name}_windows_amd64.zip ${name}_windows_amd64.exe steam_api64.dll
