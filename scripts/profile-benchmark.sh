#!/bin/bash

# Profile the Go benchmark to find performance bottlenecks

set -e

cd "$(dirname "$0")/.."

echo "╔══════════════════════════════════════════════════════════════════════════════╗"
echo "║                     BENCHMARK PROFILING TOOL                                 ║"
echo "╚══════════════════════════════════════════════════════════════════════════════╝"
echo ""
echo "This script will:"
echo "  1. Run CPU profiling on the benchmark"
echo "  2. Run memory profiling on the benchmark"
echo "  3. Generate reports showing where time/memory is spent"
echo ""

# Create profiles directory
mkdir -p profiles

echo "Running CPU profile..."
go test -bench=BenchmarkLargeSuite \
    -benchtime=3s \
    -cpuprofile=profiles/cpu.prof \
    -memprofile=profiles/mem.prof \
    -benchmem \
    ./less \
    > profiles/bench-output.txt 2>&1

echo "✓ Profiling complete"
echo ""

# Analyze CPU profile
echo "═══════════════════════════════════════════════════════════════════════════════"
echo "TOP 20 FUNCTIONS BY CPU TIME"
echo "═══════════════════════════════════════════════════════════════════════════════"
go tool pprof -text -nodecount=20 profiles/cpu.prof | head -30

echo ""
echo "═══════════════════════════════════════════════════════════════════════════════"
echo "TOP 20 FUNCTIONS BY MEMORY ALLOCATIONS"
echo "═══════════════════════════════════════════════════════════════════════════════"
go tool pprof -text -nodecount=20 -alloc_space profiles/mem.prof | head -30

echo ""
echo "═══════════════════════════════════════════════════════════════════════════════"
echo "ALLOCATION HOTSPOTS (by count)"
echo "═══════════════════════════════════════════════════════════════════════════════"
go tool pprof -text -nodecount=20 -alloc_objects profiles/mem.prof | head -30

echo ""
echo "═══════════════════════════════════════════════════════════════════════════════"
echo "INTERACTIVE ANALYSIS"
echo "═══════════════════════════════════════════════════════════════════════════════"
echo ""
echo "Profile files saved to profiles/ directory:"
echo "  • profiles/cpu.prof  - CPU profile"
echo "  • profiles/mem.prof  - Memory profile"
echo ""
echo "To explore interactively:"
echo "  go tool pprof profiles/cpu.prof"
echo "  go tool pprof profiles/mem.prof"
echo ""
echo "Useful pprof commands:"
echo "  top              - Show top functions"
echo "  top -cum         - Show top functions by cumulative time"
echo "  list <function>  - Show source code for function"
echo "  web              - Open graphical view (requires graphviz)"
echo "  traces           - Show call traces"
echo ""
echo "To generate flamegraph:"
echo "  go tool pprof -http=:8080 profiles/cpu.prof"
echo ""
