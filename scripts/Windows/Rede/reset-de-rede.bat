@echo off
:: @nome      Reset de rede completo
:: @descricao winsock, pilha TCP/IP, renova o IP e limpa o DNS
:: @admin     sim
:: @confirmar sim

echo Resetando o winsock...
netsh winsock reset
echo.
echo Resetando a pilha TCP/IP...
netsh int ip reset
echo.
echo Renovando o IP...
ipconfig /release >nul
ipconfig /renew
ipconfig /flushdns
echo.
echo Reinicie a máquina para concluir o reset do winsock.
