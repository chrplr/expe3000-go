!include "MUI2.nsh"

Name "expe3000-go"
OutFile "expe3000-setup.exe"
InstallDir "$PROGRAMFILES\expe3000-go"
RequestExecutionLevel admin

!define MUI_ABORTWARNING

!insertmacro MUI_PAGE_WELCOME
!insertmacro MUI_PAGE_DIRECTORY
!insertmacro MUI_PAGE_INSTFILES
!insertmacro MUI_PAGE_FINISH

!insertmacro MUI_UNPAGE_CONFIRM
!insertmacro MUI_UNPAGE_INSTFILES

!insertmacro MUI_LANGUAGE "English"

Section "Install"
    SetOutPath "$INSTDIR"
    
    # Files to include
    File "expe3000.exe"
    File "expe3000-gui.exe"
    File "README.txt"
    File "experiment.csv"
    File /r "assets"

    # Create uninstaller
    WriteUninstaller "$INSTDIR\Uninstall.exe"

    # Create shortcuts
    CreateDirectory "$SMPROGRAMS\expe3000-go"
    CreateShortcut "$SMPROGRAMS\expe3000-go\expe3000-gui.lnk" "$INSTDIR\expe3000-gui.exe"
    CreateShortcut "$SMPROGRAMS\expe3000-go\Uninstall.lnk" "$INSTDIR\Uninstall.exe"
SectionEnd

Section "Uninstall"
    Delete "$INSTDIR\expe3000.exe"
    Delete "$INSTDIR\expe3000-gui.exe"
    Delete "$INSTDIR\README.txt"
    Delete "$INSTDIR\experiment.csv"
    RMDir /r "$INSTDIR\assets"
    Delete "$INSTDIR\Uninstall.exe"

    RMDir "$INSTDIR"
    Delete "$SMPROGRAMS\expe3000-go\*.lnk"
    RMDir "$SMPROGRAMS\expe3000-go"
SectionEnd
