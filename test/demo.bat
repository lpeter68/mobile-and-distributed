cd ..\samplecode
for /l %%x in (1, 1, 3) do (
start cmd /k "samplecode"
)
cd ..\test 
..\samplecode\samplecode.exe < "init 100 nodes.txt"