# Create application directory
$AppDir = "$env:LOCALAPPDATA\Timesheetz"
New-Item -ItemType Directory -Force -Path $AppDir

# Copy the binary and config
Copy-Item "build\timesheetz-win-amd64.exe" "$AppDir\timesheetz.exe"
Copy-Item "config.json" "$AppDir\"

# Create a shortcut in the startup folder
$StartupFolder = "$env:APPDATA\Microsoft\Windows\Start Menu\Programs\Startup"
$WshShell = New-Object -ComObject WScript.Shell
$Shortcut = $WshShell.CreateShortcut("$StartupFolder\Timesheetz.lnk")
$Shortcut.TargetPath = "$AppDir\timesheetz.exe"
$Shortcut.WorkingDirectory = $AppDir
$Shortcut.Save()

Write-Host "Installation complete! Timesheetz has been installed to $AppDir and will start automatically on login." 