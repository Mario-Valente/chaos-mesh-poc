# Chaos Mesh POC

Uma Proof of Concept (POC) completa do **Chaos Mesh** - uma plataforma cloud-native para Chaos Engineering.

## 📋 O que é Chaos Mesh?

Chaos Mesh é uma plataforma open-source para simular falhas em ambientes Kubernetes, ajudando você a:
- Testar resiliência e tolerância a falhas das aplicações
- Validar estratégias de recuperação
- Melhorar confiabilidade em produção
- Entender pontos fracos da arquitetura

## 🏗️ Arquitetura da POC

```
┌─────────────┐
│  Frontend   │ (Go API - port 8080)
└──────┬──────┘
       │ (HTTP/REST)
┌──────▼──────┐
│   Backend   │ (Go Service - port 8081)
└──────┬──────┘
       │ (PostgreSQL Protocol)
┌──────▼──────┐
│ PostgreSQL  │ (Database)
└─────────────┘
```

### Componentes

1. **Frontend Service**: API de entrada que recebe requisições HTTP
2. **Backend Service**: Processa lógica e interage com o banco de dados
3. **PostgreSQL Database**: Armazena dados das requisições

Todos rodando em containers Docker, orquestrados por Kubernetes (Kind), com **Chaos Mesh** para injetar falhas.

## 🚀 Quick Start

### Pré-requisitos

- `kind` - Kubernetes local (https://kind.sigs.k8s.io/)
- `kubectl` - CLI do Kubernetes
- `helm` - Gerenciador de pacotes Kubernetes
- `docker` - Container runtime

### Instalação

```bash
# 1. Clone/Setup do projeto
cd chaos-mesh-poc

# 2. Setup completo (cria cluster, deploya serviços)
bash setup.sh

# 3. Instala Chaos Mesh
bash install-chaos-mesh.sh

# 4. Execute experimentos
bash run-experiments.sh
```

## 📚 Documentação

- **[Setup Guide](docs/setup.md)** - Instruções detalhadas de instalação
- **[Experiments](docs/experiments.md)** - Descrição de cada chaos experiment
- **[Observations](docs/observations.md)** - O que observar e como interpretar

## 🎯 Tipos de Chaos Experiments

### Network Chaos
- **Latency**: Adiciona delay na comunicação entre serviços
- **Packet Loss**: Simula perda de pacotes (instabilidade de rede)
- **Bandwidth Throttling**: Limita a largura de banda disponível
- **Network Partition**: Isola completamente um serviço

### Resource Chaos
- **CPU Stress**: Sobrecarrega o processador
- **Memory Stress**: Consome muita memória
- **Disk I/O Stress**: Saturação de I/O no disco

### Pod Chaos
- **Pod Kill**: Mata pods (simula crash da aplicação)
- **Pod Failure**: Marca pods como falhados
- **Container Kill**: Mata containers específicos

## 📊 Monitoramento

Após iniciar experimentos, monitore com:

```bash
# Ver status dos experiments
kubectl get chaosexperiment -n chaos-poc

# Ver logs dos serviços
kubectl logs -n chaos-poc -f deployment/frontend
kubectl logs -n chaos-poc -f deployment/backend
kubectl logs -n chaos-poc -f deployment/postgres

# Acessar dashboard do Chaos Mesh
kubectl port-forward -n chaos-mesh svc/chaos-dashboard 2333:2333
# Abra: http://localhost:2333

# Testar a API
kubectl port-forward -n chaos-poc svc/frontend-service 8080:8080
curl http://localhost:8080/api/data
```

## 🧪 Exemplo de Uso

```bash
# Terminal 1: Inicie um tester que faz requisições contínuas
kubectl port-forward -n chaos-poc svc/frontend-service 8080:8080 &
watch -n 1 'curl -s http://localhost:8080/api/data | jq .'

# Terminal 2: Execute um experimento
bash run-experiments.sh
# Escolha: 1 (Network Chaos)

# Observe os efeitos:
# - Aumento de latência
# - Timeouts
# - Erros de conexão
# - Recuperação automática após o experimento
```

## 🔧 Customização

### Modificar recursos dos containers
Edite `k8s/*-deployment.yaml` e ajuste as seções `resources:`

### Adicionar novos serviços
1. Crie novo serviço em `services/<nome>/main.go`
2. Adicione Dockerfile em `docker/<nome>.Dockerfile`
3. Crie deployment em `k8s/<nome>-deployment.yaml`
4. Deploy com `kubectl apply -f k8s/<nome>-deployment.yaml`

### Criar novo chaos experiment
1. Crie novo arquivo YAML em `chaos-experiments/`
2. Defina o CRD apropriado (NetworkChaos, StressChaos, PodChaos)
3. Aplique com `kubectl apply -f chaos-experiments/<seu-experimento>.yaml`

## 📁 Estrutura do Projeto

```
chaos-mesh-poc/
├── docker/                    # Dockerfiles
├── services/                  # Código-fonte (Go)
│   ├── frontend/
│   └── backend/
├── k8s/                       # Manifiestos Kubernetes
├── chaos-experiments/         # CRDs do Chaos Mesh
├── docs/                      # Documentação
├── setup.sh                   # Setup automático
├── install-chaos-mesh.sh      # Instala Chaos Mesh
├── run-experiments.sh         # Executa experiments
└── README.md                  # Este arquivo
```

## 🧠 O que Você Aprenderá

- Como funcionam falhas em sistemas distribuídos
- Resiliência e tolerância a falhas
- Importance de timeouts e circuit breakers
- Como recuperar de falhas automaticamente
- Observabilidade e logging em Kubernetes
- Gerenciamento de containers e orquestração

## 🔗 Recursos Úteis

- [Chaos Mesh Docs](https://chaos-mesh.org/docs/)
- [Kind Documentation](https://kind.sigs.k8s.io/docs/)
- [Kubernetes Docs](https://kubernetes.io/docs/)
- [Helm Docs](https://helm.sh/docs/)

## ⚠️ Notas

- Esta POC é para **aprendizado** em ambientes locais
- Use com cuidado em ambientes compartilhados
- Sempre tenha um plano de recuperação antes de injetar falhas
- Comece com experimentos de curta duração (2-5 minutos)

## 📝 Licença

Esta POC é fornecida "como está" para fins educacionais.

---

Divirta-se explorando Chaos Engineering! 🎉
