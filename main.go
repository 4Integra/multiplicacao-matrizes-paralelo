package main

import (
	"fmt"
	"math/rand"
	"time"

	mpi "github.com/mvneves/gompi"
)

const N = 5000

func main() {

	mpi.Init()
	defer mpi.Finalize()

	world := mpi.NewComm(true)
	rank := world.GetRank()
	size := world.GetSize()

	rowsPerProc := N / size
	rest := N % size

	localRows := rowsPerProc
	if rank < rest {
		localRows++
	}

	// quantidade de elementos locais
	localSize := localRows * N

	var A []float64
	var B []float64
	var C []float64

	var sendCounts []int
	var displs []int

	if rank == 0 {

		fmt.Printf("Multiplicação MPI %dx%d usando %d processos\n", N, N, size)

		src := rand.NewSource(42)
		rng := rand.New(src)

		A = make([]float64, N*N)
		B = make([]float64, N*N)
		C = make([]float64, N*N)

		for i := range A {
			A[i] = rng.Float64()
		}

		for i := range B {
			B[i] = rng.Float64()
		}

		sendCounts = make([]int, size)
		displs = make([]int, size)

		offset := 0
		for p := 0; p < size; p++ {

			rows := rowsPerProc
			if p < rest {
				rows++
			}

			sendCounts[p] = rows * N
			displs[p] = offset
			offset += rows * N
		}

		fmt.Println("Matrizes geradas.")
	}

	// Todos precisam de B inteira
	const tagB = 1

	if rank == 0 {

		for p := 1; p < size; p++ {
			world.Send(B, p, tagB)
		}

	} else {

		B = make([]float64, N*N)
		world.Recv(&B, 0, tagB)

	}

	const tagA = 2

	localA := make([]float64, localSize)
	localC := make([]float64, localSize)

	if rank == 0 {

		// Copia sua própria parte
		copy(localA, A[displs[0]:displs[0]+sendCounts[0]])

		// Envia as demais partes
		for p := 1; p < size; p++ {

			begin := displs[p]
			end := begin + sendCounts[p]

			world.Send(A[begin:end], p, tagA)
		}

	} else {

		world.Recv(&localA, 0, tagA)

	}

	world.Barrier()

	start := time.Now()

	// Triple loop original
	for i := 0; i < localRows; i++ {

		for j := 0; j < N; j++ {

			sum := 0.0

			for k := 0; k < N; k++ {

				sum += localA[i*N+k] * B[k*N+j]

			}

			localC[i*N+j] = sum
		}
	}

	world.Barrier()

	const tagC = 3

	if rank == 0 {

		// Copia sua parte
		copy(C[displs[0]:displs[0]+sendCounts[0]], localC)

		// Recebe dos demais processos
		for p := 1; p < size; p++ {

			temp := make([]float64, sendCounts[p])

			world.Recv(&temp, p, tagC)

			copy(C[displs[p]:displs[p]+sendCounts[p]], temp)
		}

	} else {

		world.Send(localC, 0, tagC)

	}

	if rank == 0 {

		elapsed := time.Since(start)

		fmt.Printf("\nTempo total: %v\n", elapsed)

		fmt.Printf("\nVerificação:\n")
		fmt.Printf("C[0][0] = %.15f\n", C[0])
		fmt.Printf("C[0][N-1] = %.15f\n", C[N-1])
		fmt.Printf("C[N-1][0] = %.15f\n", C[(N-1)*N])
		fmt.Printf("C[N-1][N-1] = %.15f\n", C[(N-1)*N+(N-1)])

		checksum := 0.0
		for _, v := range C {
			checksum += v
		}

		fmt.Printf("Checksum = %.15f\n", checksum)
	}
}
