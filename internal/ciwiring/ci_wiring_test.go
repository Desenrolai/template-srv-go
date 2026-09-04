// Package ciwiring não tem código de produção: existe só para fixar a fiação dos
// gatilhos do CI.
//
// O defeito que fixa: com `push` em toda branch e `pull_request` declarados
// juntos, cada commit de uma branch com PR aberto dispara DOIS runs completos do
// mesmo SHA. Medido em repo gerado a partir de um destes templates: o mesmo SHA
// com um run `success` e outro `failure`, criados com 4s de diferença. É o que
// faz alguém concluir "flaky" e ignorar o vermelho.
//
// O `concurrency` do workflow NÃO protege, e é aí que a leitura engana: a chave é
// `github.ref`, que vale `refs/heads/<branch>` no push e `refs/pull/<n>/merge` no
// pull_request. Grupos diferentes, zero cancelamento.
package ciwiring

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// stripYAMLComments remove as linhas de comentário do YAML.
//
// Obrigatório aqui: o comentário do próprio workflow documenta o defeito e cita
// o gatilho antigo textualmente. Sem esta limpeza, a prosa que explica a correção
// seria lida como a volta do defeito — comentário virando achado.
func stripYAMLComments(yaml string) string {
	linhas := strings.Split(yaml, "\n")
	limpas := make([]string, 0, len(linhas))
	for _, linha := range linhas {
		if strings.HasPrefix(strings.TrimLeft(linha, " \t"), "#") {
			continue
		}
		limpas = append(limpas, linha)
	}
	return strings.Join(limpas, "\n")
}

var (
	pushSoNaDefault = regexp.MustCompile(`push:\s*\n\s*branches:\s*\[main\]`)
	pushTodaBranch  = regexp.MustCompile(`push:\s*\n\s*branches:\s*\[['"]\*\*['"]\]`)
	prTodaBranch    = regexp.MustCompile(`pull_request:\s*\n\s*branches:\s*\[['"]\*\*['"]\]`)
	inicioOn        = regexp.MustCompile(`(?m)^on:`)
	inicioConc      = regexp.MustCompile(`(?m)^concurrency:`)
)

func lerWorkflow(t *testing.T) string {
	t.Helper()
	bruto, err := os.ReadFile(filepath.Join("..", "..", ".github", "workflows", "ci.yml"))
	if err != nil {
		t.Fatalf("não consegui ler o workflow: %v", err)
	}
	return stripYAMLComments(string(bruto))
}

// gatilhos recorta o bloco `on:` até `concurrency:`, falhando o teste se o
// recorte não casar — sem isto, os testes abaixo passariam sobre string vazia.
func gatilhos(t *testing.T, workflow string) string {
	t.Helper()
	ini := inicioOn.FindStringIndex(workflow)
	fim := inicioConc.FindStringIndex(workflow)
	if ini == nil || fim == nil || fim[0] <= ini[0] {
		t.Fatalf("não recortei o bloco de gatilhos: on=%v concurrency=%v", ini, fim)
	}
	bloco := workflow[ini[0]:fim[0]]
	if !strings.Contains(bloco, "push:") {
		t.Fatalf("bloco de gatilhos sem `push:`: %q", bloco)
	}
	return bloco
}

func TestPushRodaSoNaBranchDefault(t *testing.T) {
	if bloco := gatilhos(t, lerWorkflow(t)); !pushSoNaDefault.MatchString(bloco) {
		t.Errorf("push deveria rodar só em [main]; bloco:\n%s", bloco)
	}
}

func TestPullRequestCobreTodaBranchInclusiveFork(t *testing.T) {
	if bloco := gatilhos(t, lerWorkflow(t)); !prTodaBranch.MatchString(bloco) {
		t.Errorf("pull_request deveria cobrir toda branch; bloco:\n%s", bloco)
	}
}

func TestPushEmTodaBranchNaoVolta(t *testing.T) {
	if bloco := gatilhos(t, lerWorkflow(t)); pushTodaBranch.MatchString(bloco) {
		t.Errorf("push em toda branch duplica o run do mesmo SHA; bloco:\n%s", bloco)
	}
}

// Um teste que o CI não executa é a classe de defeito que este arquivo barra.
func TestOCIExecutaASuiteQueContemEsteTeste(t *testing.T) {
	if workflow := lerWorkflow(t); !strings.Contains(workflow, "go test ./...") {
		t.Error("o workflow deixou de rodar `go test ./...`")
	}
}

func TestRemocaoDeComentarioNaoComeConteudo(t *testing.T) {
	if got := stripYAMLComments("# nota\n  # indentado\nrun: go test\n"); got != "run: go test\n" {
		t.Errorf("comeu conteúdo: %q", got)
	}
	if got := stripYAMLComments("run: echo \"a # b\"\n"); !strings.Contains(got, "a # b") {
		t.Errorf("comeu `#` no meio da linha: %q", got)
	}
}
