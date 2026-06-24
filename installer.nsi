; NSIS Installer Script for Media Manager
; This script creates a Windows installer for the Media Manager application

!include "MUI2.nsh"

; General Configuration
Name "Media Manager"
OutFile "media-manager_installer.exe"
Unicode True
InstallDir "$PROGRAMFILES\Media Manager"
InstallDirRegKey HKCU "Software\Media Manager" ""
RequestExecutionLevel admin
Icon "internal\ui\assets\logo.ico"
UninstallIcon "internal\ui\assets\logo.ico"

; Modern UI Configuration
!define MUI_ABORTWARNING

; Pages
!insertmacro MUI_PAGE_WELCOME
!insertmacro MUI_PAGE_LICENSE "LICENSE"
!insertmacro MUI_PAGE_DIRECTORY
!insertmacro MUI_PAGE_INSTFILES
!insertmacro MUI_PAGE_FINISH

!insertmacro MUI_UNPAGE_WELCOME
!insertmacro MUI_UNPAGE_CONFIRM
!insertmacro MUI_UNPAGE_INSTFILES
!insertmacro MUI_UNPAGE_FINISH

; Languages
!insertmacro MUI_LANGUAGE "English"

; Version Information
VIProductVersion "1.0.0.0"
VIAddVersionKey "ProductName" "Media Manager"
VIAddVersionKey "CompanyName" "Joshua Hicks"
VIAddVersionKey "FileVersion" "1.0.0.0"
VIAddVersionKey "ProductVersion" "1.0.0.0"
VIAddVersionKey "FileDescription" "Media Manager Desktop Application"

; Installer Sections
Section "Media Manager" SecApp
    SectionIn RO

    SetOutPath "$INSTDIR"

    ; Copy application files
    DetailPrint "Installing Media Manager..."
    File "media-manager.exe"
    File "media-manager-service.exe"
    File "internal\ui\assets\logo.ico"

    ; Create desktop shortcut
    CreateShortCut "$DESKTOP\Media Manager.lnk" "$INSTDIR\media-manager.exe" "" "$INSTDIR\logo.ico" 0

    ; Create start menu entries
    CreateDirectory "$SMPROGRAMS\Media Manager"
    CreateShortCut "$SMPROGRAMS\Media Manager\Media Manager.lnk" "$INSTDIR\media-manager.exe" "" "$INSTDIR\logo.ico" 0
    CreateShortCut "$SMPROGRAMS\Media Manager\Media Manager Service.lnk" "$INSTDIR\media-manager-service.exe" "" "$INSTDIR\logo.ico" 0
    CreateShortCut "$SMPROGRAMS\Media Manager\Uninstall.lnk" "$INSTDIR\Uninstall.exe" "" "$INSTDIR\Uninstall.exe" 0

    ; Store installation folder
    WriteRegStr HKCU "Software\Media Manager" "" $INSTDIR

    ; Create uninstaller
    WriteUninstaller "$INSTDIR\Uninstall.exe"
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\Media Manager" "DisplayName" "Media Manager"
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\Media Manager" "UninstallString" "$INSTDIR\Uninstall.exe"
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\Media Manager" "DisplayVersion" "1.0.0"
    WriteRegStr HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\Media Manager" "Publisher" "Joshua Hicks"
    WriteRegDWord HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\Media Manager" "NoModify" 1
    WriteRegDWord HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\Media Manager" "NoRepair" 1

    ; Add Explorer context menu entries for files and folders.
    ; %1 is replaced by the selected file/folder path.
    WriteRegStr HKCR "*\shell\MediaManager.Open" "" "Open in Media Manager"
    WriteRegStr HKCR "*\shell\MediaManager.Open" "Icon" "$INSTDIR\media-manager.exe"
    WriteRegStr HKCR "*\shell\MediaManager.Open" "MultiSelectModel" "Player"
    WriteRegStr HKCR "*\shell\MediaManager.Open\command" "" "$\"$INSTDIR\media-manager.exe$\" $\"%*$\""

    WriteRegStr HKCR "*\shell\MediaManager.OpenParent" "" "Open Parent Folder in Media Manager"
    WriteRegStr HKCR "*\shell\MediaManager.OpenParent" "Icon" "$INSTDIR\media-manager.exe"
    WriteRegStr HKCR "*\shell\MediaManager.OpenParent" "MultiSelectModel" "Player"
    WriteRegStr HKCR "*\shell\MediaManager.OpenParent\command" "" "$\"$INSTDIR\media-manager.exe$\" --open-parent $\"%*$\""

    WriteRegStr HKCR "Directory\shell\MediaManager.Open" "" "Open in Media Manager"
    WriteRegStr HKCR "Directory\shell\MediaManager.Open" "Icon" "$INSTDIR\media-manager.exe"
    WriteRegStr HKCR "Directory\shell\MediaManager.Open\command" "" "$\"$INSTDIR\media-manager.exe$\" $\"%1$\""

    WriteRegStr HKCR "Directory\Background\shell\MediaManager.Open" "" "Open in Media Manager"
    WriteRegStr HKCR "Directory\Background\shell\MediaManager.Open" "Icon" "$INSTDIR\media-manager.exe"
    WriteRegStr HKCR "Directory\Background\shell\MediaManager.Open\command" "" "$\"$INSTDIR\media-manager.exe$\" $\"%V$\""

    WriteRegStr HKCR "Drive\shell\MediaManager.Open" "" "Open in Media Manager"
    WriteRegStr HKCR "Drive\shell\MediaManager.Open" "Icon" "$INSTDIR\media-manager.exe"
    WriteRegStr HKCR "Drive\shell\MediaManager.Open\command" "" "$\"$INSTDIR\media-manager.exe$\" $\"%1$\""

SectionEnd

; Uninstaller Section
Section "Uninstall"

    ; Remove files
    Delete "$INSTDIR\media-manager.exe"
    Delete "$INSTDIR\media-manager-service.exe"
    Delete "$INSTDIR\logo.ico"
    Delete "$INSTDIR\Uninstall.exe"

    ; Remove shortcuts
    Delete "$DESKTOP\Media Manager.lnk"
    Delete "$SMPROGRAMS\Media Manager\Media Manager.lnk"
    Delete "$SMPROGRAMS\Media Manager\Media Manager Service.lnk"
    Delete "$SMPROGRAMS\Media Manager\Uninstall.lnk"
    RMDir "$SMPROGRAMS\Media Manager"

    ; Remove registry entries
    DeleteRegKey HKCU "Software\Media Manager"
    DeleteRegKey HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\Media Manager"
    DeleteRegKey HKCR "*\shell\MediaManager.Open"
    DeleteRegKey HKCR "*\shell\MediaManager.OpenParent"
    DeleteRegKey HKCR "Directory\shell\MediaManager.Open"
    DeleteRegKey HKCR "Directory\Background\shell\MediaManager.Open"
    DeleteRegKey HKCR "Drive\shell\MediaManager.Open"

    ; Remove installation directory
    RMDir "$INSTDIR"

SectionEnd