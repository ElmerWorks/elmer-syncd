
[#ifndef MyAppExe]
	#define MyAppExe "elmer-syncd.exe"
[#/ifndef]

[Setup]
AppName=ElmerSyncd
AppVersion=1.0
DefaultDirName={pf64}\ElmerSyncd
DefaultGroupName=ElmerSyncd
DisableProgramGroupPage=yes

// Use preprocessor define MyAppExe to pass the built executable path when compiling:
// ISCC.exe /DMyAppExe="C:\path\to\elmer-syncd.exe" ElmerSyncdInstaller.iss

[Files]
Source: "{#MyAppExe}"; DestDir: "{app}"; Flags: ignoreversion
Source: "postinstall-message.ps1"; DestDir: "{app}"; Flags: ignoreversion

[Icons]
Name: "{group}\ElmerSyncd"; Filename: "{app}\elmer-syncd.exe"

[Run]
; Optionally run the app after install
Filename: "{app}\elmer-syncd.exe"; Description: "Start elmer-syncd"; Flags: shellexec postinstall skipifsilent
; Open the connection-info.txt in Notepad so the user sees the info in a simple UI
Filename: "notepad.exe"; Parameters: """{app}\connection-info.txt"""; Description: "Show connection information"; Flags: shellexec postinstall

[Code]
var
	RootPage: TInputDirWizardPage;

procedure InitializeWizard();
begin
	RootPage := CreateInputDirPage(wpSelectDir, 'Select Sync Root', 'Choose folder to sync to Android', 'Select the folder on this PC that will be exposed to Android clients for editing.', False, '');
	RootPage.Add('Sync root folder:');
	RootPage.Values[0] := ExpandConstant('{userdocs}');
end;

procedure CurStepChanged(CurStep: TSetupStep);
var
	rootPath: string;
	iniPath: string;
	infoPath: string;
	s: string;
begin
	if CurStep = ssPostInstall then
	begin
		rootPath := RootPage.Values[0];
		iniPath := ExpandConstant('{app}\elmer-syncd.ini');
		infoPath := ExpandConstant('{app}\connection-info.txt');
		if not DirExists(ExpandConstant('{app}')) then
			if not ForceDirectories(ExpandConstant('{app}')) then
				; // ignore failure; writing will fail later

		s := 'root=' + rootPath + #13#10;
		if not SaveStringToFile(iniPath, s, False) then
			MsgBox('Warning: could not write configuration to ' + iniPath, mbInformation, MB_OK);

		// Write a simple connection-info.txt (user can run --show-ip for LAN addresses)
		s := 'ElmerSyncd installed.' + #13#10 +
				 'Sync root: ' + rootPath + #13#10 +
				 'Examples (port 12345):' + #13#10 +
				 '  WebSocket: ws://127.0.0.1:12345/ws' + #13#10 +
				 '  REST list: http://127.0.0.1:12345/api/list?path=' + #13#10 +
				 '' + #13#10 +
				 'To get LAN IP addresses run: elmer-syncd.exe --show-ip' + #13#10 +
				 'If you configured a token, include it in header X-API-Token when connecting.';
		if not SaveStringToFile(infoPath, s, False) then
			MsgBox('Warning: could not write connection info to ' + infoPath, mbInformation, MB_OK);
	end;
end;
