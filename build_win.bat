
set GOOS=windows
set GOPACH=amd64

cd test
del lua.exe
del luac.exe
cd ..

go build -o ./test/lua.exe main/lua.go
go build -o ./test/luac.exe main/luac.go

pause