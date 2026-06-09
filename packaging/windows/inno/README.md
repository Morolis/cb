# Inno Setup (Local Build)

Download [Inno Setup](https://jrsoftware.org/isinfo.php) and extract to this directory.

Required files:
```
packaging/windows/inno/
  ISCC.exe          # The compiler
  ISCC.dll          # Compiler DLL
  Default.isl       # Default language file
  ...               # Other files from the Inno Setup distribution
```

GitHub Actions CI installs Inno Setup automatically via `choco install innosetup`.
Local builds use this directory if present.
