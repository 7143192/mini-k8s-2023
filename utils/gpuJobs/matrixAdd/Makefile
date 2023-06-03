SOURCE = matrixAdd.cu

TARGETBIN := ./matrixAdd

CC = nvcc

$(TARGETBIN):$(SOURCE)
	$(CC)  $(SOURCE) -o $(TARGETBIN)

run: $(TARGETBIN)
	$(TARGETBIN)
