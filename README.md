# Desafio Multithreading - Go

Esta aplicação em Go consulta concorrentemente duas APIs de CEP diferentes ([`BrasilAPI`](https://brasilapi.com.br) e [`ViaCEP`](https://viacep.com.br)), aceita a resposta mais rápida, descarta a mais lenta e impõe um limite estrito de 1 segundo (timeout).

## Requisitos

- Go 1.18+

## Como Executar

1. Clone ou navegue para o diretório do projeto:
   ```bash
   cd /home/jhonas/Desktop/full-cycle/desafio-multithreading
   ```

2. Execute a aplicação passando um CEP válido como argumento:
   ```bash
   go run [`main.go`](main.go:1) 01153000
   ```

## Exemplo de Saída

```text
=== Resposta mais rápida recebida de: BrasilAPI ===
CEP:          01153000
Logradouro:   Rua Vitorino Carmilo
Bairro:       Barra Funda
Cidade:       São Paulo
Estado:       SP
==========================================
```
