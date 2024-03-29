package pessoa

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/matheusCalaca/golanggrpexample/app/interface/rpc/api/pessoa"
	"github.com/matheusCalaca/golanggrpexample/app/repository"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
)

// versão da api
const (
	apiVersion = "pessoa"
)

// pessoaServiceService implementação de pessoa.PessoaServiceServer interface proto
type pessoaServiceService struct {
	db *sql.DB
}

/**
* CriarTelefone inserindo telefone no banco de dados para pesso
 */
func (service *pessoaServiceService) CriarTelefone(ctx context.Context, req *pessoa.CriarTelefoneRequest) (*pessoa.CriarTelefoneResponse, error) {
	conn, err := service.connect(ctx)
	if err != nil {
		return nil, status.Error(codes.Unknown, "Erro ao conectar ao banco de dados -> "+err.Error())
	}

	telefone := req.Telefone
	response, err := repository.CriarTelefoneRepository(conn, telefone, ctx)
	if err != nil {
		return nil, status.Error(codes.Unknown, "Erro ao Inserir Telefone ao banco de dados -> "+err.Error())
	}

	return response, nil
}

//// NewPessoaServiceServer Cria o servidor para pessoa
func NewPessoaServiceServer(db *sql.DB) pessoa.PessoaServiceServer {
	return &pessoaServiceService{db: db}
}

// checkAPI verifica se a versão da api do cliente e suportada pelo o servidor
func (service *pessoaServiceService) checkAPI(api string) error {

	if len(api) > 0 {
		if apiVersion != api {
			return status.Errorf(codes.Unimplemented,
				"unsupported API version: service implements API version '%s', but asked for '%s'", apiVersion, api)
		}
	}
	return nil
}

// connect retorna o pool de conexao com o database
func (service *pessoaServiceService) connect(ctx context.Context) (*sql.Conn, error) {
	c, err := service.db.Conn(ctx)
	if err != nil {
		return nil, status.Error(codes.Unknown, "failed to connect to database-> "+err.Error())
	}
	return c, nil
}

// Create nova pessoa
func (service *pessoaServiceService) Criar(ctx context.Context, req *pessoa.CriarPessoaRequest) (*pessoa.CriarPessoaResponse, error) {
	// check verifica se a versão da api do cliente e suportada pelo o servidor
	if err := service.checkAPI(req.Api); err != nil {
		return nil, err
	}

	reqPessoa := req.Pessoa
	// obtem a conexao com o BD
	con, err := service.connect(ctx)
	if err != nil {
		return nil, err
	}
	defer con.Close()

	// start conexão com o server
	connGrpc, err := grpc.Dial("localhost:9090", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("não foi possivel conectar : %v", err)
	}
	defer connGrpc.Close()

	// adicionando endereço para a pessoa
	pessoa.NewEnderecoServiceClient(connGrpc)

	//Endereco cliente
	enderecoResponse, err := clienteEndereco(connGrpc, err, ctx, reqPessoa.Endereco[0])
	if err != nil {
		return nil, status.Error(codes.Unknown, err.Error())
	}

	//Identidficador cliente
	identificadorResponse, err := clientIdentificador(connGrpc, ctx, reqPessoa.Identificador[0])
	if err != nil {
		return nil, status.Error(codes.Unknown, err.Error())
	}
	log.Printf("Identificador criado <%+v>\n\n", identificadorResponse.Cpf)

	telefoneResponse, err := clienteTelefone(connGrpc, ctx, reqPessoa.Telefone[0])

	log.Printf("Telefone criado <%+v>\n\n", telefoneResponse.Id)

	return repository.CriarPessoaRepository(con, ctx, reqPessoa, enderecoResponse, identificadorResponse, telefoneResponse)
}

func (service *pessoaServiceService) CriarIdentificador(ctx context.Context, req *pessoa.CriarIdentificadorRequest) (*pessoa.CriarIdentificadorResponse, error) {
	identificador := req.Identificador

	// obtem a conexao com o BD
	con, err := service.connect(ctx)
	if err != nil {
		return nil, err
	}
	defer con.Close()
	// inclusão no banco do identificador
	response, err := repository.CriarIdentificadorRepository(con, identificador, ctx)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// ----------------------- CLIENTS --------------------------------------------------

func clienteEndereco(conn *grpc.ClientConn, err error, ctx context.Context, endereco *pessoa.Endereco) (*pessoa.CriarEnderecoResponse, error) {
	// Endereço
	clientEndereco := pessoa.NewEnderecoServiceClient(conn)
	reqEndereco := pessoa.CriarEnderecoRequest{
		Api:      apiVersion,
		Endereco: endereco,
	}
	fmt.Println(reqEndereco)
	responseEndereco, err := clientEndereco.CriarEndereco(ctx, &reqEndereco)
	if err != nil {
		log.Fatalf("falha ao criar enderreço %v", err)
		return nil, status.Error(codes.Unknown, "Erro ao criar Endereço ->  "+err.Error())
	}

	return responseEndereco, nil
}

//clientIdentificador cliente para inserir um identificado na pessoa
func clientIdentificador(grpCon *grpc.ClientConn, ctx context.Context, identificador *pessoa.Identificador) (*pessoa.CriarIdentificadorResponse, error) {
	clientIdentificador := pessoa.NewPessoaServiceClient(grpCon)

	request := pessoa.CriarIdentificadorRequest{Identificador: identificador, Api: apiVersion}

	response, err := clientIdentificador.CriarIdentificador(ctx, &request)
	if err != nil {
		return nil, status.Error(codes.Unknown, "Erro ao criar o identificador -> "+err.Error())
	}
	log.Printf("Identificador criado <%+v>\n\n", request)

	return response, nil
}

// clienteTelefone cliente para inserir um telefone na tabela
func clienteTelefone(grpConn *grpc.ClientConn, ctx context.Context, telefone *pessoa.Telefone) (*pessoa.CriarTelefoneResponse, error) {

	client := pessoa.NewPessoaServiceClient(grpConn)

	request := pessoa.CriarTelefoneRequest{
		Telefone: telefone,
		Api:      apiVersion,
	}

	response, err := client.CriarTelefone(ctx, &request)
	if err != nil {
		return nil, status.Error(codes.Unknown, "Erro ao criar o Telefone -> "+err.Error())
	}

	return response, nil
}
