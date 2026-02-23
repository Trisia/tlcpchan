@echo off
rmdir /s /q target 2>nul
mkdir target
cd tlcpchan
go build -o ../target/tlcpchan.exe 
cd ..
cd tlcpchan-cli
go build -o ../target/tlcpchan-cli.exe
cd ..
cd tlcpchan-ui
call npm run build
xcopy /e /i /y ui ..\target\ui
cd ..
