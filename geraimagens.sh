#!/bin/bash

# Trabalha apenas no diretorio onde o script esta (e subdiretorios)
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

FILE="/tmp/plantuml-1.2025.10.jar"
URL="https://github.com/plantuml/plantuml/releases/download/v1.2025.10/plantuml-1.2025.10.jar"

if [ -f "$FILE" ]; then
    echo "Plantuml disponivel."
else
    echo "O arquivo $FILE nao existe. Baixando..."
    if command -v curl >/dev/null 2>&1; then
        curl -L -o "$FILE" "$URL"
    elif command -v wget >/dev/null 2>&1; then
        wget -O "$FILE" "$URL"
    else
        echo "Nenhum downloader (curl/wget) disponivel." >&2
        exit 1
    fi
    echo "Download concluido."
fi

echo "Produzindo arquivos SVG a partir de arquivos PlantUML (recursivamente). Aguarde ..."

regenerated=0
skipped=0
while IFS= read -r -d '' puml; do
    dir=$(dirname "$puml")
    outdir="$dir/imagens"
    mkdir -p "$outdir"

    basename_puml=$(basename "$puml" .puml)
    svg_file="$outdir/$basename_puml.svg"

    if [ ! -f "$svg_file" ] || [ "$puml" -nt "$svg_file" ]; then
        java -jar "$FILE" -quiet -tsvg -nometadata --output-dir "$outdir" "$puml"
        echo "  [GERADO] $puml"
        regenerated=$((regenerated+1))
    else
        echo "  [SKIP]   $puml (sem alteracoes)"
        skipped=$((skipped+1))
    fi
done < <(find . -type f -name "*.puml" -not -path "*/imagens/*" -print0)

echo "Geracao concluida: $regenerated arquivo(s) gerado(s), $skipped ignorado(s)."
