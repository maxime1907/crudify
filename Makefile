NAME		=	crudify

GOCMD		=	go

GORUN		=	$(GOCMD) run
GOBUILD		=	$(GOCMD) build
GOINSTALL 	=	$(GOCMD) install
GOCLEAN		=	$(GOCMD) clean
GOTEST		=	$(GOCMD) test
GOGET		=	$(GOCMD) get

all: libs

libs:
	$(GOINSTALL) .

tests:
	$(GOTEST) ./config/
	$(GOTEST) ./dbhelper/
	$(GOTEST) ./handler/
	$(GOTEST) ./logger/
	$(GOTEST) ./router/
	$(GOTEST) ./auth/

func:
	$(GOTEST) ./test/

bench:
	$(GOTEST) ./benchmark/ -v

clean:
	$(GOCLEAN)
	rm -f $(NAME)
