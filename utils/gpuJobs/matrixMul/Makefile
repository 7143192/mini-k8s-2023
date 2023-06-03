SOURCE = matrixMul.cu

TARGETBIN := ./matrixMul

CC = nvcc

$(TARGETBIN):$(SOURCE)
	$(CC)  $(SOURCE) -o $(TARGETBIN)

run: $(TARGETBIN)
	$(TARGETBIN)
