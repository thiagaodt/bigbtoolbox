# @nome      Informações do sistema
# @descricao SO, processador, memória e discos
# @so        windows

$os  = Get-CimInstance Win32_OperatingSystem
$cpu = Get-CimInstance Win32_Processor | Select-Object -First 1
$pc  = Get-CimInstance Win32_ComputerSystem

[pscustomobject]@{
    'Máquina'      = $env:COMPUTERNAME
    'Sistema'      = '{0} ({1})' -f $os.Caption, $os.Version
    'Processador'  = $cpu.Name.Trim()
    'Memória'      = '{0:N1} GB' -f ($pc.TotalPhysicalMemory / 1GB)
    'Ligado desde' = $os.LastBootUpTime
} | Format-List

Get-CimInstance Win32_LogicalDisk -Filter 'DriveType=3' |
    Select-Object @{ n = 'Unidade'; e = { $_.DeviceID } },
                  @{ n = 'Tamanho (GB)'; e = { [math]::Round($_.Size / 1GB, 1) } },
                  @{ n = 'Livre (GB)'; e = { [math]::Round($_.FreeSpace / 1GB, 1) } } |
    Format-Table -AutoSize
