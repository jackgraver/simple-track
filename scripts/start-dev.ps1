$ProjectRoot = "C:\Users\Jacks Desktop\Desktop\Coding\simpletracker"
$BackendDir = Join-Path $ProjectRoot "backend"
$FrontendDir = Join-Path $ProjectRoot "frontend"
$PostgresContainer = "simpletracker-postgres-dev-1"
$DockerDesktop = Join-Path ${env:ProgramFiles} "Docker\Docker\Docker Desktop.exe"

function Test-DockerRunning {
    docker info 2>$null | Out-Null
    return $LASTEXITCODE -eq 0
}

function Start-DockerDesktop {
    if (-not (Test-Path $DockerDesktop)) {
        Write-Error "Docker Desktop not found at $DockerDesktop"
        exit 1
    }
    Write-Host "Starting Docker Desktop..."
    Start-Process $DockerDesktop
    Write-Host "Waiting for Docker to be ready..."
    for ($i = 0; $i -lt 60; $i++) {
        Start-Sleep -Seconds 2
        if (Test-DockerRunning) {
            Write-Host "Docker is ready."
            return
        }
    }
    Write-Error "Docker did not become ready within 2 minutes."
    exit 1
}

function Start-Postgres {
    $exists = docker inspect $PostgresContainer 2>$null
    if ($LASTEXITCODE -ne 0) {
        Write-Host "Creating postgres container via docker compose..."
        Push-Location $ProjectRoot
        docker compose up -d postgres-dev
        Pop-Location
        return
    }
    $running = docker inspect -f '{{.State.Running}}' $PostgresContainer 2>$null
    if ($running -eq 'true') {
        Write-Host "Postgres container already running."
        return
    }
    Write-Host "Starting postgres container..."
    docker start $PostgresContainer
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Failed to start $PostgresContainer"
        exit 1
    }
}

function Start-DevTerminal {
    param(
        [string]$Title,
        [string]$Directory,
        [string]$Command
    )
    $args = "/k title $Title && cd /d `"$Directory`" && $Command"
    Start-Process cmd -ArgumentList $args -WindowStyle Normal
}

if (-not (Test-DockerRunning)) {
    Start-DockerDesktop
} else {
    Write-Host "Docker is already running."
}

Start-Postgres

Write-Host "Starting backend (air)..."
Start-DevTerminal -Title "SimpleTracker Backend" -Directory $BackendDir -Command "air"

Write-Host "Starting frontend (npm run dev)..."
Start-DevTerminal -Title "SimpleTracker Frontend" -Directory $FrontendDir -Command "npm run dev"

Write-Host ""
Write-Host "Dev environment started."
Write-Host "  Postgres: localhost:5432 (simpletracker_dev)"
Write-Host "  Backend:  air in new terminal"
Write-Host "  Frontend: npm run dev in new terminal"
