$ErrorActionPreference = 'Stop'
$projectRoot = Split-Path -Parent $PSScriptRoot
$mediaPath = Join-Path $projectRoot 'assets\demo'

docker exec utilities-minio mkdir -p /tmp/utilities-demo
if ($LASTEXITCODE -ne 0) { throw 'MinIO is unavailable. Run docker compose up -d first.' }
docker cp "$mediaPath\." utilities-minio:/tmp/utilities-demo/
if ($LASTEXITCODE -ne 0) { throw 'Cannot copy media to MinIO.' }
docker exec utilities-minio sh -c 'mc alias set demo http://localhost:9000 "$MINIO_ROOT_USER" "$MINIO_ROOT_PASSWORD" >/dev/null && mc mb --ignore-existing demo/utilities >/dev/null && mc anonymous set download demo/utilities >/dev/null && mc cp --recursive /tmp/utilities-demo/ demo/utilities/'
if ($LASTEXITCODE -ne 0) { throw 'Cannot upload media to the bucket.' }

$sql = @'
UPDATE utilities SET image_url = 'http://localhost:9000/utilities/' || CASE name
 WHEN 'Электроэнергия' THEN 'electricity.webp'
 WHEN 'Отопление' THEN 'heating.webp'
 WHEN 'Холодное водоснабжение' THEN 'cold-water.avif'
 WHEN 'Газоснабжение' THEN 'gas-supply.avif'
 WHEN 'Вывоз ТКО' THEN 'waste-removal.jpg'
 WHEN 'Капитальный ремонт' THEN 'capital-repair.webp'
 END
WHERE name IN ('Электроэнергия','Отопление','Холодное водоснабжение','Газоснабжение','Вывоз ТКО','Капитальный ремонт');
'@
$tempSql = Join-Path ([IO.Path]::GetTempPath()) ('utilities-media-' + [guid]::NewGuid() + '.sql')
try {
    [IO.File]::WriteAllText($tempSql, $sql, (New-Object Text.UTF8Encoding $false))
    docker cp $tempSql utilities-postgres:/tmp/utilities-media.sql
    if ($LASTEXITCODE -ne 0) { throw 'Cannot copy the SQL file.' }
    docker exec utilities-postgres psql -v ON_ERROR_STOP=1 -U postgres -d utilities -f /tmp/utilities-media.sql
    if ($LASTEXITCODE -ne 0) { throw 'Cannot update media URLs.' }
} finally {
    if (Test-Path -LiteralPath $tempSql) { Remove-Item -LiteralPath $tempSql }
}
Write-Host 'Demo media uploaded. Database image URLs updated; other data preserved.'
