# e-vendas — Funcionalidades e requisitos do produto

Documento de referência do produto. Descreve **o que** o e-vendas faz, para quem e
com quais regras. O **como** (código, arquitetura, tecnologias) fica por sua conta.

**Prioridades usadas neste documento**

| Marca | Significado |
|---|---|
| **MVP** | Entra na primeira versão que vai para as lojas-piloto. |
| **Fase 2** | Entra depois que o MVP estiver sendo usado de verdade. |
| **Futuro** | Boa ideia, mas só faz sentido com o produto já crescendo. |

---

## 1. Visão do produto

**O que é.** Uma plataforma de vendas online para o comércio **local** da sua cidade
e região: lojas, lanchonetes e pessoas que vendem por conta própria. Cada vendedor
ganha uma loja online com catálogo, carrinho e pedidos. O cliente escolhe os
produtos, finaliza o pedido e a loja recebe tudo organizado, com o pedido também
chegando pelo WhatsApp.

**Problema que resolve.** Hoje o pequeno vendedor atende por WhatsApp e Instagram:
manda foto, combina preço, anota pedido no caderno, perde mensagem e não sabe
quanto vendeu. O cliente não sabe o que a loja tem nem se está aberta.

**Proposta de valor**

- **Para o lojista:** loja online pronta em poucos minutos, sem site, sem
  complicação, com pedidos organizados e um link para divulgar.
- **Para o comprador:** achar tudo o que a cidade vende em um só lugar, pedir pelo
  celular em poucos toques e acompanhar o pedido.
- **Para a cidade:** o dinheiro e a divulgação ficam no comércio local.

**Princípios do produto**

1. **Celular primeiro.** Quase todo comprador e boa parte dos lojistas vão usar
   pelo celular.
2. **Simples para quem não é de tecnologia.** O lojista precisa cadastrar o
   primeiro produto sem ler manual.
3. **WhatsApp é parte do fluxo, não concorrente.** O pedido chega organizado no
   sistema e também avisa no WhatsApp, onde o vendedor já vive.
4. **Pix e pagamento na entrega primeiro.** Pagamento online dentro da plataforma
   só depois que houver CNPJ e volume.
5. **Pouco para entregar bem.** Melhor poucas funções funcionando do que muitas
   pela metade.

---

## 2. Perfis de usuário

| Perfil | Quem é | O que precisa fazer |
|---|---|---|
| **Visitante / comprador** | Morador da cidade. Não precisa criar conta. | Achar lojas e produtos, montar carrinho, pedir, acompanhar. |
| **Lojista (dono)** | Dono da loja, lanchonete ou vendedor. | Criar a loja, cadastrar produtos, receber e tocar pedidos, ver vendas. |
| **Funcionário da loja** | Pessoa que ajuda a atender. | Ver e atualizar pedidos. *(Fase 2)* |
| **Administrador da plataforma** | Você. | Aprovar lojas, cuidar de categorias, acompanhar a plataforma. |

---

## 3. Fluxos principais

### 3.1 Comprador faz um pedido
1. Abre o link da loja (compartilhado no WhatsApp/Instagram) ou entra na home da cidade.
2. Navega pelo catálogo, abre um produto e adiciona ao carrinho.
3. Abre o carrinho, confere itens e quantidades.
4. Informa nome e telefone, escolhe **retirada** ou **entrega** (com endereço) e a
   **forma de pagamento**.
5. Confere o resumo (subtotal, taxa de entrega, total) e envia o pedido.
6. O sistema registra o pedido, mostra a confirmação com o número do pedido e abre
   o WhatsApp com a mensagem pronta para a loja.
7. Acompanha o status pelo link de acompanhamento.

### 3.2 Lojista começa a vender
1. Cria a conta (nome, e-mail, senha) e aceita os termos.
2. Preenche os dados da loja: nome, WhatsApp, endereço, horário, retirada/entrega,
   formas de pagamento.
3. Cadastra os primeiros produtos com foto, preço e categoria.
4. A loja fica **pendente de aprovação** e, aprovada, passa a aparecer na cidade.
5. Divulga o link e o QR Code da loja.

### 3.3 Lojista toca um pedido
1. Recebe o alerta de novo pedido (aviso na tela e e-mail; WhatsApp do próprio
   pedido).
2. Abre o pedido, confere itens, endereço e pagamento.
3. **Confirma** o pedido (ou cancela com motivo).
4. Avança: em preparo → pronto (retirada) ou saiu para entrega → entregue.
5. Marca como **pago** quando receber.
6. A cada passo, pode chamar o cliente no WhatsApp com uma mensagem pronta.

### 3.4 Administrador aprova uma loja
1. Vê a lista de lojas novas pendentes.
2. Confere dados e produtos (nome, fotos, conteúdo permitido).
3. Aprova ou recusa, com motivo. A loja aprovada passa a ser pública.

---

## 4. Funcionalidades por módulo

### A. Contas e acesso

| ID | Funcionalidade | Descrição | Prioridade |
|---|---|---|---|
| A-01 | Cadastro do lojista | Nome, e-mail, senha e aceite dos termos de uso e da política de privacidade. | MVP |
| A-02 | Login e logout | Entrar com e-mail e senha; sair em qualquer tela. | MVP |
| A-03 | Recuperar senha | Link de redefinição enviado por e-mail, com validade curta. | MVP |
| A-04 | Confirmação de e-mail | Garante que o e-mail é real antes de publicar a loja. | MVP |
| A-05 | Meu perfil | Alterar nome, e-mail e senha. | MVP |
| A-06 | Proteção de acesso | Bloqueio temporário após várias tentativas erradas; sessão que expira. | MVP |
| A-07 | Equipe da loja | Convidar funcionários com permissão limitada (só pedidos). | Fase 2 |
| A-08 | Login com Google | Entrar sem digitar senha. | Fase 2 |
| A-09 | Conta do comprador | Histórico de pedidos e endereços salvos. | Fase 2 |

### B. Loja

| ID | Funcionalidade | Descrição | Prioridade |
|---|---|---|---|
| B-01 | Dados da loja | Nome, descrição curta, WhatsApp, endereço, bairro, cidade e estado. | MVP |
| B-02 | Endereço web da loja | Nome na URL escolhido pelo lojista (ex.: o nome da loja sem espaços), único na plataforma. | MVP |
| B-03 | Logo | Imagem da loja que aparece no cabeçalho e nas listas. | MVP |
| B-04 | Horário de funcionamento | Horário de abertura e fechamento por dia da semana. | MVP |
| B-05 | Pausar pedidos | Botão para fechar a loja na hora (acabou o estoque, feriado, imprevisto) sem mexer no horário. | MVP |
| B-06 | Retirada e entrega | A loja escolhe se aceita retirada, entrega ou os dois. Pelo menos um é obrigatório. | MVP |
| B-07 | Taxa de entrega | Valor fixo de entrega definido pela loja. | MVP |
| B-08 | Pedido mínimo | Valor mínimo para a loja aceitar o pedido. | MVP |
| B-09 | Formas de pagamento aceitas | Pix, dinheiro e cartão na entrega. Para Pix, a chave da loja e a instrução de pagamento. Para dinheiro, pergunta se precisa de troco. | MVP |
| B-10 | Tempo estimado | Estimativa de preparo e de entrega exibida ao cliente. | MVP |
| B-11 | Situação da loja | Pendente de aprovação, ativa ou suspensa; o lojista vê a situação e o motivo, se recusada. | MVP |
| B-12 | Capa da loja | Imagem grande no topo da página da loja. | Fase 2 |
| B-13 | Redes sociais | Link do Instagram e outras redes. | Fase 2 |
| B-14 | Taxa de entrega por bairro | Valores diferentes por região. | Fase 2 |
| B-15 | Entrega por distância | Cálculo automático pela distância. | Futuro |

### C. Catálogo de produtos

| ID | Funcionalidade | Descrição | Prioridade |
|---|---|---|---|
| C-01 | Cadastrar, editar e desativar produto | Nome, descrição, preço, categoria. Produto vendido nunca some: é desativado. | MVP |
| C-02 | Fotos do produto | Até 5 fotos, escolher a principal e a ordem. Redução automática de tamanho. | MVP |
| C-03 | Categoria | Escolher uma categoria da lista da plataforma. | MVP |
| C-04 | Disponível ou esgotado | Liga e desliga o produto na vitrine com um toque. | MVP |
| C-05 | Estoque opcional | A loja pode informar a quantidade; sem informar, não há controle. | MVP |
| C-06 | Código do produto (SKU) | Código interno opcional, único dentro da loja. | MVP |
| C-07 | Observação por item | O cliente escreve um recado no item (ex.: "sem cebola", "embrulhar para presente"). | MVP |
| C-08 | Adicionais e opções | Grupos como "Escolha o molho" ou "Adicionais" com preço extra; essencial para lanchonetes. | Fase 2 |
| C-09 | Variações | Tamanho e cor com estoque próprio; essencial para roupas. | Fase 2 |
| C-10 | Preço promocional | Mostra "de R$ X por R$ Y". | Fase 2 |
| C-11 | Duplicar produto | Cria cópia para agilizar cadastro de itens parecidos. | Fase 2 |
| C-12 | Seções do cardápio da loja | A loja organiza os produtos em seções próprias (ex.: "Combos", "Bebidas"). | Fase 2 |
| C-13 | Produtos em destaque | A loja escolhe os que aparecem primeiro. | Fase 2 |
| C-14 | Importar por planilha | Cadastro em massa. | Futuro |

### D. Vitrine pública (o que o comprador vê)

| ID | Funcionalidade | Descrição | Prioridade |
|---|---|---|---|
| D-01 | Home da cidade | Lojas da cidade, categorias em destaque e campo de busca. | MVP |
| D-02 | Lista por categoria | Todas as lojas e produtos de uma categoria. | MVP |
| D-03 | Busca | Procura por nome de loja ou de produto. | MVP |
| D-04 | Página da loja | Logo, nome, aberta ou fechada, horário, endereço, tempo estimado, taxa, pedido mínimo e catálogo. | MVP |
| D-05 | Página do produto | Fotos, descrição, preço, observação, quantidade e botão de adicionar. | MVP |
| D-06 | Compartilhar | Compartilhar loja ou produto por link, WhatsApp e cópia do endereço. | MVP |
| D-07 | Prévia nos links | Ao colar o link no WhatsApp ou Instagram, aparece foto, nome e descrição. | MVP |
| D-08 | Aparecer no Google | Páginas de loja e produto com endereços amigáveis e títulos próprios. | MVP |
| D-09 | Filtros | Aberto agora, entrega, retirada, ordenar por nome. | Fase 2 |
| D-10 | Lojas em destaque e banners | Espaço de divulgação na home. | Fase 2 |
| D-11 | Avaliações | O cliente avalia a loja depois do pedido entregue. | Fase 2 |
| D-12 | Favoritos | Salvar lojas preferidas. | Futuro |

### E. Carrinho e finalização do pedido

| ID | Funcionalidade | Descrição | Prioridade |
|---|---|---|---|
| E-01 | Carrinho por loja | O carrinho é sempre de uma loja. Se o cliente tentar adicionar item de outra loja, o sistema avisa e pergunta se quer esvaziar o carrinho. | MVP |
| E-02 | Editar carrinho | Alterar quantidades, remover item, editar observação. | MVP |
| E-03 | Dados do cliente | Nome e telefone (WhatsApp). Sem senha e sem cadastro. | MVP |
| E-04 | Retirada ou entrega | Mostra só o que a loja aceita. | MVP |
| E-05 | Endereço de entrega | Rua, número, bairro, complemento e ponto de referência. | MVP |
| E-06 | Forma de pagamento | Pix, dinheiro (com troco) ou cartão na entrega, conforme a loja aceita. | MVP |
| E-07 | Resumo do pedido | Itens, subtotal, taxa de entrega e total. | MVP |
| E-08 | Enviar pedido | Registra o pedido e abre o WhatsApp da loja com a mensagem pronta. | MVP |
| E-09 | Confirmação | Tela com número do pedido, resumo e instrução de pagamento (chave Pix, se for o caso). | MVP |
| E-10 | Acompanhar pedido | Link próprio, difícil de adivinhar, mostrando o status atual. | MVP |
| E-11 | Validações | Bloqueia loja fechada, pedido abaixo do mínimo, produto esgotado ou sem estoque, campos faltando. | MVP |
| E-12 | Lembrar dados | Guarda nome, telefone e endereço neste aparelho para o próximo pedido. | MVP |
| E-13 | Preencher endereço pelo CEP | Completa rua e bairro automaticamente. | Fase 2 |
| E-14 | Pix com confirmação automática | Pagamento dentro da plataforma, com baixa automática. | Fase 2 |
| E-15 | Cupom de desconto | Código promocional da loja. | Fase 2 |
| E-16 | Agendar pedido | Escolher dia e hora, inclusive com a loja fechada. | Fase 2 |
| E-17 | Cartão online | Pagamento por cartão dentro da plataforma. | Futuro |

### F. Gestão de pedidos (lojista)

| ID | Funcionalidade | Descrição | Prioridade |
|---|---|---|---|
| F-01 | Lista de pedidos | Pedidos separados por situação (novos, em preparo, prontos, finalizados), com busca por número, nome ou telefone. | MVP |
| F-02 | Detalhe do pedido | Cliente, itens com observações, endereço, forma de pagamento, valores e histórico de mudanças. | MVP |
| F-03 | Mudar situação | Avançar o pedido pelas etapas permitidas. | MVP |
| F-04 | Marcar como pago | Confirma o recebimento (Pix, dinheiro ou cartão). | MVP |
| F-05 | Cancelar com motivo | Registra o motivo; devolve o estoque, se houver controle. | MVP |
| F-06 | Chamar no WhatsApp | Abre a conversa com o cliente com mensagem pronta para cada situação ("Seu pedido saiu para entrega"). | MVP |
| F-07 | Alerta de novo pedido | Som e aviso destacado enquanto o painel está aberto, mais e-mail. | MVP |
| F-08 | Imprimir comanda | Versão do pedido própria para impressão no navegador. | MVP |
| F-09 | Anotação interna | Nota que só a loja vê. | Fase 2 |
| F-10 | Notificação no celular | Aviso mesmo com o painel fechado. | Fase 2 |
| F-11 | Histórico do cliente | Ver pedidos anteriores do mesmo telefone. | Fase 2 |
| F-12 | Impressora térmica | Impressão automática de comanda. | Futuro |

### G. Resumo e relatórios

| ID | Funcionalidade | Descrição | Prioridade |
|---|---|---|---|
| G-01 | Resumo do dia | Pedidos novos, em andamento, finalizados, faturamento do dia e atalhos. | MVP |
| G-02 | Vendas por período | Faturamento e quantidade de pedidos por dia, semana e mês. | Fase 2 |
| G-03 | Produtos mais vendidos | Ranking por quantidade e por valor. | Fase 2 |
| G-04 | Ticket médio e clientes que voltam | Indicadores de saúde da loja. | Fase 2 |
| G-05 | Visitas e cliques no WhatsApp | Quantas pessoas viram a loja e quantas pediram. | Fase 2 |
| G-06 | Exportar relatórios | Planilha para o contador. | Fase 2 |

### H. Divulgação

| ID | Funcionalidade | Descrição | Prioridade |
|---|---|---|---|
| H-01 | Link da loja | Endereço curto para copiar e colar na bio do Instagram e no WhatsApp. | MVP |
| H-02 | QR Code da loja | Imagem para baixar e usar em cartaz, balcão e embalagem. | MVP |
| H-03 | Texto pronto | Mensagem de divulgação já escrita para colar no WhatsApp. | MVP |
| H-04 | Cartaz para imprimir | Arte pronta com nome da loja e QR Code. | Fase 2 |

### I. Administração da plataforma

| ID | Funcionalidade | Descrição | Prioridade |
|---|---|---|---|
| I-01 | Acesso de administrador | Área separada, protegida, só para você. | MVP |
| I-02 | Lojas | Lista com busca e filtro por situação; aprovar, recusar (com motivo), suspender e reativar. | MVP |
| I-03 | Categorias | Criar, renomear, ordenar, ativar e desativar categorias e subcategorias. | MVP |
| I-04 | Visão geral | Lojas ativas, pedidos por dia, lojas novas aguardando, lojas sem venda. | MVP |
| I-05 | Usuários | Buscar usuário, desativar conta e forçar redefinição de senha. | MVP |
| I-06 | Entrar como a loja (suporte) | Ver o painel como o lojista para ajudar, com registro de que isso aconteceu. | Fase 2 |
| I-07 | Denúncias e moderação | Cliente denuncia loja ou produto; você analisa. | Fase 2 |
| I-08 | Destaques | Controlar lojas em destaque e banners da home. | Fase 2 |
| I-09 | Registro de auditoria | Histórico de quem fez o quê no painel administrativo. | Fase 2 |

### J. Planos e cobrança da plataforma (o seu faturamento)

| ID | Funcionalidade | Descrição | Prioridade |
|---|---|---|---|
| J-01 | Período grátis | Lojas-piloto sem pagar, com data de término visível. | Fase 2 |
| J-02 | Planos e mensalidade | Plano gratuito limitado e plano pago, com cobrança recorrente (Pix ou cartão). | Fase 2 |
| J-03 | Bloqueio por inadimplência | Aviso e depois suspensão suave da loja. | Fase 2 |
| J-04 | Comissão por venda | Percentual sobre cada pedido pago pela plataforma. | Futuro |

### K. Comunicação e notificações

| ID | Funcionalidade | Descrição | Prioridade |
|---|---|---|---|
| K-01 | E-mails automáticos | Boas-vindas, confirmar e-mail, recuperar senha, novo pedido, loja aprovada ou recusada. | MVP |
| K-02 | Mensagens por link do WhatsApp | O sistema monta o texto e abre o WhatsApp; a pessoa só toca em enviar. Sem custo e sem API oficial. | MVP |
| K-03 | Aviso no navegador | Notificação no celular ou computador mesmo com a aba fechada. | Fase 2 |
| K-04 | WhatsApp automático | Mensagens enviadas sozinhas pela API oficial do WhatsApp Business. | Futuro |

---

## 5. Regras de negócio

1. **Um pedido é de uma única loja.** Carrinho com produtos de lojas diferentes não existe: o cliente finaliza uma loja por vez.
2. **O pedido congela preço e nome.** Se o lojista mudar o preço depois, pedidos antigos não mudam.
3. **Loja fechada não recebe pedido.** Vale para fora do horário, loja pausada, pendente de aprovação ou suspensa. A vitrine mostra "fechada" e bloqueia a finalização.
4. **Produto desativado ou esgotado não pode ser comprado.** Com estoque controlado, não se vende acima do saldo; o saldo baixa quando o pedido é confirmado e volta se for cancelado.
5. **Pedido mínimo e taxa de entrega** são aplicados no resumo antes de enviar.
6. **Total do pedido** = subtotal − desconto + taxa de entrega.
7. **Entrega exige endereço completo; retirada mostra o endereço da loja.**
8. **Etapas do pedido:** novo → confirmado → em preparo → pronto → (saiu para entrega, só se for entrega) → entregue. **Cancelado** pode acontecer em qualquer etapa antes de entregue. Entregue e cancelado são finais.
9. **Só a loja (e o administrador) muda a situação do pedido.** O cliente apenas acompanha.
10. **Pagamento no MVP é combinado com a loja.** Quem marca "pago" é o lojista. Pix usa a chave da própria loja.
11. **O cliente é identificado pelo telefone.** Não há senha. O link de acompanhamento é secreto e não pode ser adivinhado.
12. **Número do pedido é sequencial por loja** (#0001, #0002...), para facilitar a conversa entre loja e cliente.
13. **Loja nova fica pendente até a aprovação.** Decisão em aberto: aprovar manualmente todas ou liberar as lojas-piloto automaticamente.
14. **Categorias são controladas pelo administrador.** O lojista só escolhe.
15. **Nada que já foi vendido é apagado.** Produtos e lojas com pedidos são desativados.
16. **Cada loja só vê e altera os próprios dados.** Dados de clientes são visíveis apenas para a loja do pedido.
17. **Endereço web da loja é único e permanente.** Mudar exige cuidado, porque o link já foi divulgado.
18. **Imagens:** aceitar JPG, PNG e WebP, com tamanho máximo e redução automática.
19. **Uma loja precisa aceitar retirada ou entrega.** Não pode ficar sem nenhuma.

---

## 6. Requisitos não funcionais

**Experiência**
- Funciona bem em celular de entrada e internet lenta; carregamento rápido das páginas públicas.
- Botões grandes, textos legíveis, contraste adequado, navegação por teclado e leitor de tela nos pontos principais.
- Tudo em português do Brasil: moeda em reais, datas no formato brasileiro, horário local da cidade.
- Mensagens de erro claras e que dizem como resolver ("Falta o telefone do cliente").

**Segurança**
- Senhas guardadas de forma protegida, nunca em texto legível.
- Toda comunicação criptografada (HTTPS).
- Cada loja isolada das outras: nenhum lojista consegue ver ou mexer em dado de outro.
- Proteção contra abuso: limite de tentativas de login, de pedidos por minuto e de envios de imagem.
- Validação de tudo que o usuário digita ou envia, inclusive arquivos.
- Link de acompanhamento do pedido e de redefinição de senha impossíveis de adivinhar.
- Nunca guardar dados de cartão.

**Privacidade (LGPD)**
- Termos de uso e política de privacidade publicados, com aceite no cadastro do lojista e aviso claro no pedido do comprador.
- Coletar só o necessário: nome, telefone e endereço de entrega; CPF e e-mail só se forem realmente precisos.
- Permitir que a pessoa peça acesso, correção ou exclusão dos seus dados.
- Definir por quanto tempo os dados de pedidos são guardados.
- Cuidado extra com dados de menores se o produto vier a atender esse público.

**Confiabilidade**
- Cópias de segurança automáticas do banco, com teste de restauração.
- Registro de erros e aviso quando algo quebrar.
- Imagens guardadas fora do servidor da aplicação.
- Plano simples para quando o sistema cair (página de aviso e canal de contato).

**Busca e divulgação**
- Páginas de loja e produto indexáveis, com títulos e descrições próprios, e prévia bonita ao compartilhar.

**Desempenho e custo**
- Páginas públicas rápidas e que aguentam um pico (por exemplo, divulgação em grupo de WhatsApp da cidade).
- Manter o custo de infraestrutura baixo no início.

---

## 7. Aspectos legais e administrativos

> Não substitui orientação de contador ou advogado. Vale uma conversa antes de cobrar.

- **CNPJ e enquadramento:** necessário antes de cobrar dos lojistas. Nem toda atividade de software cabe em todos os regimes; confirme com um contador.
- **Termos de uso e política de privacidade:** obrigatórios antes de abrir ao público.
- **Responsabilidades:** deixar claro nos termos que o lojista responde pelo produto, preço, entrega, nota fiscal e atendimento; a plataforma só oferece a ferramenta.
- **Produtos proibidos:** definir e publicar a lista (ilegais, armas, medicamentos controlados, bebida alcoólica para menores etc.).
- **Marca e domínio:** conferir disponibilidade do nome, do domínio e do perfil nas redes, e consultar o INPI.
- **Pagamentos de terceiros:** ao oferecer pagamento dentro da plataforma, o gateway escolhido traz exigências próprias de cadastro e de repasse.
- **Nota fiscal:** no MVP, emissão fica por conta do lojista fora da plataforma.

---

## 8. Impacto no banco de dados

O modelo atual (10 tabelas) cobre o núcleo: usuários, lojas, equipe, categorias,
produtos, fotos, clientes, pedidos e itens. Para sustentar as funcionalidades
**MVP** acima, ele precisa destes ajustes. São listas de intenção, sem código.

**Ajustes nas tabelas existentes**
- **Loja:** situação (pendente, ativa, suspensa) e motivo da recusa; pedido mínimo; tempo estimado de preparo e de entrega; chave Pix e instruções de pagamento; formas de pagamento aceitas; indicação de "pausada para novos pedidos".
- **Usuário:** indicação de que é administrador da plataforma; data de confirmação do e-mail.
- **Pedido:** número sequencial por loja; código secreto para o link de acompanhamento; valor do troco quando o pagamento é em dinheiro; motivo do cancelamento; datas de cada mudança de situação.
- **Item do pedido:** observação do cliente para aquele item.
- **Situação do pedido:** incluir o novo e as etapas de entrega conforme a regra 8 (ajustar a lista de valores aceitos).

**Tabelas novas para o MVP**
- **Redefinição de senha** e **confirmação de e-mail** (códigos com validade).
- **Histórico de situação do pedido** (quem mudou, de quê para quê, quando).
- **Horário de funcionamento** da loja, se for preferível a uma estrutura própria em vez de um campo livre.

**Tabelas para a Fase 2**
- Cupons, avaliações, planos e assinaturas, cobranças, notificações, registro de auditoria, opções e adicionais de produto, variações, endereços salvos do cliente, denúncias.

---

## 9. Inventário de telas

Estas são as telas do MVP. Elas são as mesmas que vão aparecer nas imagens de
referência.

### 9.1 Comprador (celular)

| ID | Tela | Funcionalidades |
|---|---|---|
| C1 | Home da cidade | D-01, D-03 |
| C2 | Busca e categorias | D-02, D-03 |
| C3 | Página da loja | D-04, D-06, D-07 |
| C4 | Página do produto | D-05, C-07 |
| C5 | Carrinho | E-01, E-02 |
| C6 | Finalizar pedido | E-03 a E-07, E-11, E-12 |
| C7 | Pedido enviado | E-08, E-09 |
| C8 | Acompanhar pedido | E-10 |

### 9.2 Lojista (computador, com versão para celular nos pedidos)

| ID | Tela | Funcionalidades |
|---|---|---|
| L1 | Entrar | A-02, A-03 |
| L2 | Criar conta | A-01, A-04 |
| L3 | Primeiros passos | B-01 a B-11, C-01 |
| L4 | Resumo do dia | G-01 |
| L5 | Pedidos (computador) | F-01, F-03, F-07 |
| L5m | Pedidos (celular) | F-01, F-03, F-07 |
| L6 | Detalhe do pedido | F-02, F-04, F-05, F-06, F-08 |
| L7 | Produtos | C-01, C-04, C-05 |
| L8 | Cadastro de produto | C-01 a C-07 |
| L9 | Minha loja | B-01 a B-11 |
| L10 | Divulgação | H-01 a H-03 |

### 9.3 Administrador (computador)

| ID | Tela | Funcionalidades |
|---|---|---|
| A1 | Lojas | I-02 |
| A2 | Categorias | I-03 |
| A3 | Visão geral | I-04 |

**Fora das imagens de referência (Fase 2):** relatórios, planos e cobrança, equipe,
avaliações, cupons, adicionais e variações de produto.

---

## 10. Roadmap sugerido

Os prazos são estimativas para uma pessoa trabalhando sozinha e dependem das horas
disponíveis.

| Etapa | Objetivo | O que entrega |
|---|---|---|
| **0. Validação** | Confirmar que os lojistas querem. | Conversa com 10 a 15 vendedores; 5 ou 6 lojas-piloto combinadas; nome, domínio e perfil nas redes reservados. |
| **1. Base** | Lojista consegue existir no sistema. | Contas, login, loja, categorias, produtos com fotos, aprovação pelo administrador. |
| **2. Vitrine** | Comprador consegue ver e escolher. | Home da cidade, página da loja, produto, busca, links compartilháveis. |
| **3. Pedido** | A venda acontece. | Carrinho, finalização, WhatsApp, acompanhamento, gestão de pedidos, alertas. |
| **4. Piloto** | Aprender com uso real. | 5 a 10 lojas usando, ajustes semanais, resumo do dia, divulgação (link e QR Code). |
| **5. Abertura** | Crescer na cidade. | Divulgação ao público, CNPJ, termos publicados, início da cobrança. |
| **6. Fase 2** | Reter e ganhar mais. | Adicionais e variações, Pix automático, relatórios, equipe, avaliações, cupons, planos. |

---

## 11. Como medir se está dando certo

- **Lojas ativas:** quantas publicaram pelo menos um produto.
- **Tempo até o primeiro produto:** quanto leva do cadastro à primeira publicação.
- **Pedidos por semana**, no total e por loja.
- **Lojas com pelo menos um pedido** nos primeiros 14 dias.
- **Lojas que continuam** depois de 30 e de 90 dias.
- **Pedidos concluídos / pedidos cancelados.**
- **Visitas que viram pedido** (conversão da vitrine).

---

## 12. Riscos e como reduzir

| Risco | Como reduzir |
|---|---|
| Lojistas não adotam. | Validar antes de construir; cadastrar a primeira leva de produtos junto com eles; onboarding de 15 minutos. |
| Compradores não aparecem (poucas lojas, poucos clientes). | Lançar com lojas combinadas antes; divulgar em grupos da cidade; link e QR Code nas próprias lojas. |
| Lojista acha trabalhoso cadastrar produtos. | Cadastro rápido pelo celular, foto primeiro, poucos campos; ajudar no começo. |
| Pedido se perde ou lojista não vê. | Alerta forte no painel, e-mail e WhatsApp pronto no pedido. |
| Problemas de segurança ou vazamento de dados. | Isolamento entre lojas, senhas protegidas, limites de abuso, LGPD desde o início. |
| Tentar fazer tudo sozinho e travar. | Manter o MVP enxuto e cortar o que não é essencial para a primeira venda. |
| Custo de infraestrutura crescer antes da receita. | Começar com hospedagem simples e barata; monitorar o gasto. |

---

## 13. Decisões em aberto

1. **Nome do produto:** o repositório se chama "Loja" e o produto, e-vendas. Qual fica?
2. **Endereço das lojas:** subdomínio por loja ou caminho no mesmo site?
3. **Aprovação de lojas:** manual para todas ou automática para as lojas-piloto?
4. **Cobrança:** mensalidade fixa, comissão por venda ou plano grátis com limite? Qual valor?
5. **Cidade e região do piloto:** só a sua cidade no início ou já as vizinhas?
6. **Entrega:** só as lojas entregando por conta própria, ou parceria com motoboys?
7. **O comprador tem conta?** No MVP não precisa, mas vale decidir quando entra.
8. **Limites do plano grátis** (número de produtos, fotos por produto), se houver.
9. **Lista de produtos e conteúdos proibidos.**
10. **Quem emite nota fiscal** e como isso aparece para o comprador.
