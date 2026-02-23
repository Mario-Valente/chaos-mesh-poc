# Observations & Insights

Notas sobre o que esperar de cada experimento e insights sobre Chaos Engineering.

## Comportamento Esperado

### Network Chaos → Latency

**Timeline esperado:**
```
T=0s: Latência normal (~100ms)
T=0.5s: Latência inicia (agora ~600ms)
T=5m: Experimento termina, volta ao normal
```

**Sinais de resiliência:**
- ✅ Frontend continua respondendo com latência aumentada
- ✅ Logs mostram tempo de resposta aumentado
- ✅ Nenhuma falha de conexão (ainda dentro do timeout de 10s)

**Sinais de problemas:**
- ❌ Timeout começam a ocorrer
- ❌ Erros "connection refused"
- ❌ Cascata de erros

**O que indica:**
Se latência de 500ms causa timeout, significa que você precisa:
- Aumentar timeout do cliente
- Implementar retry logic
- Adicionar circuit breaker

---

### Network Chaos → Packet Loss

**Esperado:**
```
T=0s: 0% de erros
T=0.5s: Erros aumentam (30% loss = ~30% de requests falhando)
T=5m: Volta ao normal
```

**Observar nos logs:**
```
[Backend] Error reading backend response: EOF
[Backend] Error calling backend: connection reset by peer
[Frontend] Error: backend unavailable
```

**Aprendizado:**
- Packet loss acontece em redes reais (especialmente mobile/WAN)
- Seu código precisa de retry logic
- Circuit breaker pode prevenir cascata de erros

---

### Pod Chaos → Pod Kill

**Timeline:**
```
T=0s: 2 pods saudáveis
T=30s: 1 pod morre, Kubernetes cria novo
T=60s: Novo pod está ready, 2 saudáveis novamente
T=2m: Experimento termina
```

**Monitorar:**
```bash
watch kubectl get pods -n chaos-poc
```

**O que observar:**
- Alguns requests falham durante morte do pod
- Load balancer (kube-proxy) redireciona para outro pod
- Novo pod é criado automaticamente
- Recuperação é rápida (segundos)

**Importante:**
- Sem múltiplas réplicas, todo serviço cai
- Com 2+ réplicas, serviço continua (degradado)
- Health checks são críticos

---

### Resource Chaos → CPU Stress

**Expectativas:**
```
Latência normal: ~100ms
Com stress: ~500-1000ms (CPU contention)
```

**Sinais:**
```
kubectl top pods -n chaos-poc
# Esperado: CPU em máximo durante stress
```

**O que aprender:**
1. CPU stress simula:
   - Outras aplicações no mesmo host
   - Code ineficiente
   - Loops infinitos

2. Soluções:
   - Caching
   - Optimization
   - Horizontal scaling
   - Resource limits e requests

---

### Pod Chaos → Database Container Kill

**Timeline:**
```
T=0s: Database saudável
T=30s: Container morre, Kubernetes reinicia
T=60s: Database está down (recovery leva tempo)
T=120s: Database pronto novamente
T=durante: Backend falha a escrever, reporta erro
```

**Diferença vs Pod Kill:**
- PostgreSQL leva mais tempo para recuperar (fsync, WAL, etc)
- Conexões ativas são perdidas
- Clientes precisam retry

**Observar:**
```
[Backend] Error writing to database: connection refused
Requests reportam: "db_error"
```

---

## 💡 Key Insights

### 1. Múltiplas Réplicas são Críticas

```
Sem HA:
❌ 1 pod morre → Serviço inteiro cai

Com HA:
✅ 1 pod morre → Serviço continua em degraded mode
✅ 2+ pods saudáveis → Recuperação rápida
```

### 2. Timeouts Salvam Vidas

```go
// Sem timeout
client := &http.Client{} // timeout infinito
// Resultado: Hang indefinido

// Com timeout apropriado
client := &http.Client{
    Timeout: 10 * time.Second,
}
// Resultado: Fail fast e permite retry
```

### 3. Circuit Breaker Pattern

```
Comportamento sem circuit breaker:
Request 1 → timeout
Request 2 → timeout
Request 3 → timeout
... cascata infinita

Comportamento com circuit breaker:
Requests 1-3 → timeout
Circuit abre (OPEN state)
Requests 4+ → rápido "circuit open" (fail fast)
Recuperação: circuit tenta reconectar (HALF-OPEN)
```

### 4. Graceful Degradation

```
Melhor: Retornar dados em cache em vez de erro
Ainda melhor: Retornar dados parciais/degradados
Pior: Erro genérico que não explica o problema
```

No nosso backend:
```go
if db == nil {
    status = "no_db" // Explora o problema
}
```

### 5. Health Checks são Essenciais

```yaml
livenessProbe:  # "Você está vivo?"
readinessProbe: # "Você consegue servir tráfego?"
```

Sem eles:
- Kubernetes mantém tráfego para pods mortos
- Requests falham desnecessariamente

---

## 📊 Comparação de Estratégias

### Frente a Latência Aumentada

| Estratégia | Vantagem | Desvantagem |
|-----------|----------|------------|
| Aumentar timeout | Simples | Usuários esperam mais |
| Retry com backoff | Recupera falhas transitórias | Pode ampliar problema |
| Circuit breaker | Fail fast | Mais complexo |
| Cache | Sem chamada backend | Dados podem estar stale |

**Melhor: Combinação de todas**

### Frente a Pod Failure

| Estratégia | Resultado |
|-----------|-----------|
| 1 pod, sem HA | Downtime total |
| 2 pods, manual recovery | Downtime até human intervir |
| 3+ pods, auto-healing | Zero downtime |

---

## 🔍 Debugging

### "Por que meu experimento não teve efeito?"

Checklist:
1. ✅ Experimento foi aplicado? `kubectl get chaosexperiment -n chaos-poc`
2. ✅ Pods alvo existem? `kubectl get pods -n chaos-poc -l app=backend`
3. ✅ Labels batem? Verificar `chaos-experiments/*.yaml` vs deployments
4. ✅ Duração suficiente? (Espere ~10s depois de aplicar)
5. ✅ Chaos daemon rodando? `kubectl get pods -n chaos-mesh`

### "Como remover um experimento?"

```bash
# Ver experimentos ativas
kubectl get chaosexperiment -n chaos-poc

# Remover uma
kubectl delete chaosexperiment -n chaos-poc <nome>

# Remover todas
kubectl delete chaosexperiment -n chaos-poc --all
```

### "Meu serviço está travado"

```bash
# Ver eventos
kubectl describe pod <pod-name> -n chaos-poc

# Ver logs recentes
kubectl logs <pod-name> -n chaos-poc --tail=50

# Logs anteriores se crashed
kubectl logs <pod-name> -n chaos-poc --previous

# Forçar restart
kubectl rollout restart deployment/<name> -n chaos-poc
```

---

## 📈 Métricas para Coletar

Quando executar experimentos em ambiente real, coletar:

1. **Request Latency**
   - P50, P95, P99
   - Max latency

2. **Error Rate**
   - Percent de requests falhados
   - Tipos de erro

3. **Throughput**
   - Requests per second
   - Bytes transferred

4. **Resource Usage**
   - CPU / Memory
   - Network I/O

5. **Recovery Time**
   - Tempo até sistema se recuperar
   - Time to restore normal operation

---

## 🎓 Resultados Esperados por Fase

### Fase 1: Network Chaos (Dias 1-2)

**Você vai aprender:**
- Como latência afeta usuários
- Importância de timeouts
- Quando retry funciona

**Esperado após:**
- Implementar timeouts apropriados
- Adicionar retry logic
- Melhorar observabilidade de rede

### Fase 2: Resource Chaos (Dias 3-4)

**Você vai aprender:**
- Limites de recursos importantes
- Autoscaling quando necessário
- Performance bottlenecks

**Esperado após:**
- Right-size dos containers
- Implementar horizontal scaling
- Identificar code ineficiencies

### Fase 3: Pod Chaos (Dias 5-6)

**Você vai aprender:**
- Redundância salva vidas
- Rápida recuperação é crítica
- Health checks são essenciais

**Esperado após:**
- 3+ replicas em produção
- Implementar circuit breakers
- Graceful shutdown logic

---

## ✅ Checklist de Learnings

Depois de completar todos os experimentos, você deve conseguir:

- [ ] Explicar o que é Chaos Engineering e por quê é importante
- [ ] Implementar timeouts apropriados em calls
- [ ] Implementar retry logic com exponential backoff
- [ ] Implementar circuit breaker pattern
- [ ] Configurar múltiplas réplicas de serviço
- [ ] Implementar health checks (liveness + readiness)
- [ ] Usar Chaos Mesh para injetar falhas
- [ ] Interpretar resultados de chaos experiments
- [ ] Escrever SLAs baseado em chaos results
- [ ] Criar runbooks de recovery

---

Parabéns por explorar Chaos Engineering! 🎉

Próximos passos:
1. Experimente combinar múltiplos chaos experiments
2. Aumente a duração dos experimentos
3. Implemente monitoramento real (Prometheus + Grafana)
4. Execute em staging antes de produção
