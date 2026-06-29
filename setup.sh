#!/bin/bash
# Configura o ambiente para compilar Go com MPI no cluster Atlântica.
# Uso: source setup.sh

export PKG_CONFIG_PATH=/LADAPPs/OpenMPI/openmpi-4.1.1/lib/pkgconfig

echo "Verificando ambiente..."
echo ""

echo "--- Go ---"
go version 2>/dev/null || echo "ERRO: Go não encontrado no PATH"
echo ""

echo "--- OpenMPI ---"
mpirun --version 2>/dev/null || echo "ERRO: mpirun não encontrado no PATH"
echo ""

echo "--- pkg-config (ompi) ---"
if pkg-config --exists ompi 2>/dev/null; then
    echo "CFLAGS: $(pkg-config --cflags ompi)"
    echo "LIBS:   $(pkg-config --libs ompi)"
else
    echo "ERRO: pkg-config não encontrou o pacote 'ompi'"
    echo "Verifique se PKG_CONFIG_PATH está correto"
fi
echo ""

echo "Ambiente configurado."
