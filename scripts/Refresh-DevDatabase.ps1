[CmdletBinding()]
param(
    [switch]$Force
)

$ErrorActionPreference = "Stop"
$ProjectRoot = Split-Path -Parent $PSScriptRoot
# $SanitizeScript = Join-Path $PSScriptRoot "sanitize-dev-db.sql"
$DumpName = "simpletracker-prod-$([guid]::NewGuid().ToString('N')).dump"
$HostDumpPath = Join-Path $env:TEMP $DumpName

function Invoke-Compose {
    param([string[]]$Arguments)
    & docker compose @Arguments
    if ($LASTEXITCODE -ne 0) {
        throw "docker compose $($Arguments -join ' ') failed."
    }
}

function Get-ComposeContainerId {
    param([string]$Service)
    $containerId = (& docker compose ps -q $Service).Trim()
    if ($LASTEXITCODE -ne 0 -or [string]::IsNullOrWhiteSpace($containerId)) {
        throw "Could not find the $Service container."
    }
    return $containerId
}

if (-not (Get-Command docker -ErrorAction SilentlyContinue)) {
    throw "Docker is required."
}

# if (-not (Test-Path $SanitizeScript)) {
#     throw "Missing sanitization script: $SanitizeScript"
# }

if (-not $Force) {
    $confirmation = Read-Host "This permanently replaces simpletracker_dev with a copy of simpletracker_prod. Type REFRESH to continue"
    if ($confirmation -cne "REFRESH") {
        Write-Host "Refresh cancelled."
        exit 0
    }
}

Push-Location $ProjectRoot
try {
    Invoke-Compose @("up", "-d", "postgres-prod", "postgres-dev")
    $sourceContainer = Get-ComposeContainerId "postgres-prod"
    $targetContainer = Get-ComposeContainerId "postgres-dev"
    $dbUser = (& docker compose exec -T postgres-prod printenv POSTGRES_USER).Trim()
    if ($LASTEXITCODE -ne 0 -or [string]::IsNullOrWhiteSpace($dbUser)) {
        throw "Could not determine POSTGRES_USER from postgres-prod."
    }

    Invoke-Compose @("exec", "-T", "postgres-prod", "pg_dump", "-U", $dbUser, "--format=custom", "--no-owner", "--no-privileges", "--file=/tmp/$DumpName", "simpletracker_prod")
    & docker cp "${sourceContainer}:/tmp/$DumpName" $HostDumpPath
    if ($LASTEXITCODE -ne 0) {
        throw "Could not copy the production dump from the container."
    }

    $terminateConnections = "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = 'simpletracker_dev' AND pid <> pg_backend_pid();"
    Invoke-Compose @("exec", "-T", "postgres-dev", "psql", "-v", "ON_ERROR_STOP=1", "-U", $dbUser, "-d", "postgres", "-c", $terminateConnections)
    Invoke-Compose @("exec", "-T", "postgres-dev", "dropdb", "-U", $dbUser, "--if-exists", "simpletracker_dev")
    Invoke-Compose @("exec", "-T", "postgres-dev", "createdb", "-U", $dbUser, "simpletracker_dev")

    & docker cp $HostDumpPath "${targetContainer}:/tmp/$DumpName"
    if ($LASTEXITCODE -ne 0) {
        throw "Could not copy the production dump into the development container."
    }
    Invoke-Compose @("exec", "-T", "postgres-dev", "pg_restore", "-U", $dbUser, "--no-owner", "--no-privileges", "--exit-on-error", "-d", "simpletracker_dev", "/tmp/$DumpName")
#    Get-Content -Raw $SanitizeScript | & docker compose exec -T postgres-dev psql "-v" "ON_ERROR_STOP=1" "-U" $dbUser "-d" "simpletracker_dev"
#    if ($LASTEXITCODE -ne 0) {
#        throw "Development database sanitization failed."
#    }

    Write-Host "Development database refreshed."
}
finally {
    if (Test-Path $HostDumpPath) {
        Remove-Item $HostDumpPath -Force
    }
    if ($sourceContainer) {
        & docker exec $sourceContainer rm -f "/tmp/$DumpName" | Out-Null
    }
    if ($targetContainer) {
        & docker exec $targetContainer rm -f "/tmp/$DumpName" | Out-Null
    }
    Pop-Location
}
