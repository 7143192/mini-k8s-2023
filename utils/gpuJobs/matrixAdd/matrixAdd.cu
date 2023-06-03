// #include <stdio.h>
// #include "cuda_runtime.h"
// #include "device_launch_parameters.h"
// #include <cuda_runtime.h>
#include <stdio.h>
#include <iostream>

using namespace std;

#define CHECK(call)                                                    \
  {                                                                    \
    const cudaError_t error = call;                                    \
    if (error != cudaSuccess) {                                        \
      printf("ERROR: %s:%d,", __FILE__, __LINE__);                     \
      printf("code:%d,reason:%s\n", error, cudaGetErrorString(error)); \
      exit(1);                                                         \
    }                                                                  \
  }

// init device.(device <--> running on GPU. )
void initDevice(int devNum) {
  int dev = devNum;
  cudaDeviceProp deviceProp;
  CHECK(cudaGetDeviceProperties(&deviceProp, dev));
  printf("Using device %d: %s\n", dev, deviceProp.name);
  CHECK(cudaSetDevice(dev));
}

// init the matrix.
void initMatrix(int* matrix, int size) {
  for (int i = 0; i < size; ++i) {
    matrix[i] = i + 1;
  }
}

// print out final result to DEBUG
//NOTE: in lyh's VM, there is NO GPU Device to debug....
void printMatrix(int* mat, int nBytes){
  printf("matrixAdd running res = \n");
  for(int i=0; i < nBytes; i++){
    printf("%d ", mat[i]);
    if (i != 0) {
      if (i % 1023 == 0) {
        printf("\n");
      }
    }
  }
}

__global__
void matrixAdd(int* A, int* B, int* res, int nx,int ny) {
  int ix = threadIdx.x + blockDim.x * blockIdx.x;
  int iy = threadIdx.y + blockDim.y * blockIdx.y;
  int idx = ix + iy * ny;
  if (ix >= nx || iy >= ny) {
    return ;
  }
  res[idx] = A[idx] + B[idx];
}

int main(int argc, char** argv) {
  //init dev
  initDevice(0);

  int nx = 1 << 10;
  int ny = 1 << 10;
  int nBytes = nx * ny * sizeof(int);
  // variables used in HOST device.
  int* A_host = (int*) malloc(nBytes);
  int* B_host = (int*) malloc(nBytes);
  int* host_res = (int*) malloc(nBytes);
  initMatrix(A_host, nx * ny);
  initMatrix(B_host, nx * ny);
  // variables used in DEVICE(GPU) side.
  int* A_dev = NULL;
  int* B_dev = NULL;
  int* dev_res = NULL;
  // use function cudaMalloc(void**, int) to malloc in GPU side.
  CHECK(cudaMalloc((void**)&A_dev, nBytes));
  CHECK(cudaMalloc((void**)&B_dev, nBytes));
  CHECK(cudaMalloc((void**)&dev_res, nBytes));
  // use function cudaMemcpy(void* dst, const void* src, size_t count, cudaMemcpyKind kind) to copy var from GPU back to HOST(CPU).
  CHECK(cudaMemcpy(A_dev, A_host, nBytes, cudaMemcpyHostToDevice));
  CHECK(cudaMemcpy(B_dev, B_host, nBytes, cudaMemcpyHostToDevice));
  // block: 16 * 16
  dim3 threadsPerBlock(16, 16);
  // NOTE: do not forget "+1" operation!
  // grid size : 64 * 64
  dim3 numBlocks((nx - 1) / threadsPerBlock.x + 1, (ny - 1) / threadsPerBlock.y + 1);
  // call the __global__ function to add every position of the two matrixes parallelly.
  matrixAdd<<<numBlocks, threadsPerBlock>>>(A_dev, B_dev, dev_res, nx, ny);
  // use the function sync to wait all functions finish.
  CHECK(cudaDeviceSynchronize());
  // copy final result of MatrixAdd from GPU to CPU.
  CHECK(cudaMemcpy(host_res, dev_res, nBytes, cudaMemcpyDeviceToHost));
  // used to DEBUG.
  printMatrix(host_res, nx * ny);

  // use function CudaFree() to free var in GPU side.
  cudaFree(A_dev);
  cudaFree(B_dev);
  cudaFree(dev_res);
  // normal free function in CPU side.
  free(A_host);
  free(B_host);
  free(host_res);
  // cudaDeviceReset();
  return 0;
}