# Chaos Experiments Guide

Descrição detalhada de cada experimento disponível na POC.

## Como Executar Experimentos

### Usar o script interativo

```bash
bash run-experiments.sh
# Escolha uma das opções (1-4)
```

### Aplicar manualmente

```bash
# Aplicar um experimento
kubectl apply -f chaos-experiments/network-chaos.yaml

# Monitorar
kubectl get chaosexperiment -n chaos-poc -w

# Remover
kubectl delete chaosexperiment -n chaos-poc --all
```

## 🌐 Network Chaos Experiments

### 1. Backend Latency (Latência)

**O que faz**: Adiciona delay de 500ms ± 100ms nas respostas do backend.

**Como observar**:
```bash
# Terminal 1: Teste contínuo
while true; do curl -w "Time: %{time_total}s\n" http://localhost:8080/api/data; sleep 1; done

# Terminal 2: Aplicar experimento
kubectl apply -f chaos-experiments/network-chaos.yaml
```

**Esperado**:
- Tempo de resposta aumenta de ~100ms para ~600ms
- Frontend continua respondendo (timeout de 10s)
- Backend inteiro está afetado

**Aprendizado**:
- Importância de timeouts apropriados
- Como latência afeta user experience
- Retry logic em clientes

---

### 2. Backend Packet Loss (Perda de Pacotes)

**O que faz**: Drop de 30% dos pacotes para o backend.

**Como observar**:
```bash
kubectl logs -n chaos-poc -f deployment/backend
# Procure por erros de conexão
```

**Esperado**:
- Aumento de erros e timeouts
- Logs mostram conexões instáveis
- Alguns requests falham completamente

**Aprendizado**:
- Resiliência em redes instáveis
- Necessidade de retry com backoff
- Circuit breakers

---

### 3. Backend Bandwidth Limit (Throttling)

**O que faz**: Limita bandwidth para 1 Mbps.

**Como observar**:
```bash
# Monitor de bytes trocados
kubectl top pod -n chaos-poc
```

**Esperado**:
- Transferência mais lenta
- Possíveis timeouts se houver muito volume

**Aprendizado**:
- Limitações de banda em redes mobile/WAN
- Cache strategies
- Compression

---

### 4. PostgreSQL Network Partition (Isolamento)

**O que faz**: Isola completamente o PostgreSQL (bloqueia todos os pacotes).

**Como observar**:
```bash
# Teste continuará respondendo, mas com erro
curl http://localhost:8080/api/data
# Resposta: erro de conexão ao banco

# Logs
kubectl logs -n chaos-poc -f deployment/backend
```

**Esperado**:
- Backend continua respondendo (HTTP)
- Mas erros ao acessar banco de dados
- Responses com status "db_error"

**Aprendizado**:
- Graceful degradation
- Fallback strategies
- Importance de health checks

---

## 💻 Resource Chaos Experiments

### 1. Backend CPU Stress (Sobrecarga de CPU)

**O que faz**: Utiliza ~80% de CPU em 2 workers no backend.

**Como observar**:
```bash
# Terminal 1: Testes
while true; do curl -w "Time: %{time_total}s\n" http://localhost:8080/api/data; sleep 1; done

# Terminal 2: Monitor
kubectl top pods -n chaos-poc --containers

# Terminal 3: Aplicar
kubectl apply -f chaos-experiments/resource-chaos.yaml
```

**Esperado**:
- Aumento de latência
- Possível throttling do container
- Kubelet pode matar pod se exceder limits

**Aprendizado**:
- Tuning de resources
- Autoscaling triggers
- Performance degradation patterns

---

### 2. Backend Memory Stress (Pressão de Memória)

**O que faz**: Aloca ~256Mi de memória.

**Como observar**:
```bash
# Monitor de memória
kubectl top pods -n chaos-poc

# Observar OOMKilled
kubectl get pods -n chaos-poc
# Se OOMKilled, pod será reiniciado
```

**Esperado**:
- Memória em máximo
- Pod pode ser OOMKilled e reiniciado
- Service continua disponível com outras réplicas

**Aprendizado**:
- Memory leaks detection
- Resource limits importance
- High availability com múltiplas réplicas

---

### 3. PostgreSQL Memory Stress

**O que faz**: Estresse de memória no banco de dados.

**Esperado**:
- Database pode ficar lento
- Queries podem timeout
- Possível crash do PostgreSQL
- Backend reporta erros

---

### 4. PostgreSQL Disk I/O Stress

**O que faz**: Simula I/O intenso no disco.

**Esperado**:
- Database queries ficam lentas
- CPU pode aumentar
- Latência nas respostas

---

## 📦 Pod Chaos Experiments

### 1. Backend Pod Kill (Matar Pods)

**O que faz**: Mata um pod do backend (com 2 réplicas).

**Como observar**:
```bash
# Terminal 1: Watch dos pods
kubectl get pods -n chaos-poc -w

# Terminal 2: Testes contínuos
while true; do curl -s http://localhost:8080/api/data | jq '.'; sleep 0.5; done

# Terminal 3: Aplicar
kubectl apply -f chaos-experiments/pod-chaos.yaml
```

**Esperado**:
- Um pod morre e é recriado
- Alguns requests falham momentaneamente
- Serviço continua disponível (replica saudável)
- Recuperação em segundos

**Aprendizado**:
- High availability com múltiplas réplicas
- Load balancing automático
- Stateless design benefits

---

### 2. Backend Pod Failure (Falha de Pods)

**O que faz**: Marca 50% dos pods como falhados.

**Como observar**:
```bash
kubectl logs -n chaos-poc -f deployment/backend
# Procure por logs de "failure injection"
```

**Esperado**:
- Comportamento similar ao pod-kill
- Maior impacto pois afeta múltiplas réplicas

---

### 3. PostgreSQL Container Kill

**O que faz**: Mata o container do PostgreSQL.

**Como observar**:
```bash
kubectl get pods -n chaos-poc postgres-* -w

# Logs
kubectl logs -n chaos-poc pod/postgres-* --previous
```

**Esperado**:
- PostgreSQL container morre
- É reiniciado automaticamente
- Backend durante isso reporta erros
- Recuperação leva mais tempo que pod chaos

**Aprendizado**:
- Database resilience
- Connection pooling
- Retry strategies para databases

---

## 📊 Interpretando Resultados

### Métricas para observar

1. **Latência**
   ```bash
   curl -w "@format.txt" http://localhost:8080/api/data
   # Ou parse do tempo_time_total
   ```

2. **Taxa de erro**
   ```bash
   # Contar erros nos logs
   kubectl logs -n chaos-poc deployment/backend | grep -i error | wc -l
   ```

3. **Disponibilidade**
   - Quantos requests foram bem-sucedidos?
   - Quanto tempo levou para recuperar?

4. **Uso de recursos**
   ```bash
   kubectl top pods -n chaos-poc --containers
   ```

## 🎯 Sugestões de Experimentos

### Iniciante
1. Comece com **Network Latency** (5 min)
2. Depois **Pod Kill** (2 min)
3. Por fim **Backend CPU Stress** (5 min)

### Intermediário
1. Combine múltiplos: Latency + Pod Kill
2. Teste com maior duration
3. Observe padrões de recuperação

### Avançado
1. Crie experimentos customizados
2. Teste com múltiplas cargas simultâneas
3. Integre com monitoring/alerting

## ⚠️ Boas Práticas

1. **Sempre comece pequeno**
   - Latência baixa (100-500ms)
   - Duração curta (2-5 min)
   - Um serviço por vez

2. **Tenha um plano de recuperação**
   - Saiba como remover experimento
   - Tenha rollback automático

3. **Monitore os resultados**
   - Logs
   - Métricas
   - Alertas

4. **Documente learnings**
   - O que funcionou
   - O que quebrou
   - Como melhorar

Veja [Observations](observations.md) para mais insights.
