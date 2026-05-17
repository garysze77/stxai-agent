// ── State ──
let currentTicker = '';
let currentLang = navigator.language.startsWith('zh')
  ? (navigator.language === 'zh-CN' ? 'zh-CN' : 'zh-HK')
  : 'en';

// ── DOM refs ──
const $ = (id) => document.getElementById(id);
const tickerForm = $('ticker-form');
const tickerInput = $('ticker-input');
const cards = $('cards');
const welcome = $('welcome');
const statusText = $('status-text');
const chatSection = $('chat-section');
const chatForm = $('chat-form');
const chatInput = $('chat-input');
const chatMessages = $('chat-messages');

// ── Language ──
document.querySelectorAll('.lang-btn').forEach(btn => {
  btn.addEventListener('click', () => {
    document.querySelectorAll('.lang-btn').forEach(b => b.classList.remove('active'));
    btn.classList.add('active');
    currentLang = btn.dataset.lang;
  });
});

// Set initial active language
const initialLangBtn = document.querySelector(`[data-lang="${currentLang}"]`);
if (initialLangBtn) {
  document.querySelectorAll('.lang-btn').forEach(b => b.classList.remove('active'));
  initialLangBtn.classList.add('active');
}

// ── Ticker search ──
tickerForm.addEventListener('submit', (e) => {
  e.preventDefault();
  const raw = tickerInput.value.trim();
  if (!raw) return;
  currentTicker = raw.toUpperCase();
  loadDashboard(currentTicker);
});

// Quick ticker buttons
document.querySelectorAll('.quick-ticker').forEach(btn => {
  btn.addEventListener('click', () => {
    currentTicker = btn.dataset.ticker;
    tickerInput.value = currentTicker;
    loadDashboard(currentTicker);
  });
});

// ── Dashboard loading ──
async function loadDashboard(ticker) {
  welcome.style.display = 'none';
  cards.style.display = 'grid';
  chatSection.style.display = 'block';
  chatMessages.innerHTML = '';
  chatInput.value = '';
  setStatus('Loading ' + ticker + '…', 'status-loading');

  // Reset all cards to loading state
  resetCards();

  // 1. Price first (fast)
  try {
    const priceData = await fetchJSON('GET', '/api/v1/price/' + ticker);
    renderPrice(priceData);
    setStatus('Price loaded — fetching indicators…', 'status-loading');
  } catch (e) {
    setStatus('Failed to load price: ' + e.message, 'status-error');
    return;
  }

  // 2. Indicators + Scan + News in parallel
  const results = await Promise.allSettled([
    fetchJSON('GET', '/api/v1/analyze/' + ticker + '?lang=' + currentLang),
    fetchJSON('POST', '/api/v1/scan', { ticker: ticker, market: detectMarket(ticker) }),
    fetchJSON('GET', '/api/v1/news/' + ticker + '?lang=' + currentLang),
  ]);

  const [analyzeResult, scanResult, newsResult] = results;

  if (analyzeResult.status === 'fulfilled') {
    renderIndicators(analyzeResult.value);
  }
  if (scanResult.status === 'fulfilled') {
    renderPatterns(scanResult.value);
  }
  if (newsResult.status === 'fulfilled') {
    renderNews(newsResult.value);
  }

  const errors = [analyzeResult, scanResult, newsResult]
    .filter(r => r.status === 'rejected')
    .map(r => r.reason.message);

  if (errors.length > 0) {
    setStatus('Some data unavailable: ' + errors.join('; '), 'status-error');
  } else {
    setStatus('Dashboard ready for ' + ticker, '');
  }
}

function resetCards() {
  $('price-value').textContent = '…';
  $('price-name').textContent = '';
  $('rsi-value').textContent = '…';
  $('rsi-zone').textContent = '';
  $('rsi-zone').className = 'card-sub';
  $('macd-value').textContent = '…';
  $('macd-sub').textContent = '';
  $('ma50-value').textContent = '…';
  $('ma200-value').textContent = '…';
  $('ma-relation').textContent = '';
  $('patterns-list').innerHTML = '<span class="muted">Scanning…</span>';
  $('bb-upper').textContent = '…';
  $('bb-middle').textContent = '…';
  $('bb-lower').textContent = '…';
  $('news-list').innerHTML = '<span class="muted">Loading news…</span>';
}

// ── Card renderers ──
function renderPrice(data) {
  if (!data || data.price == null) {
    $('price-value').textContent = 'N/A';
    return;
  }
  $('price-value').textContent = '$' + data.price.toLocaleString();
  $('price-name').textContent = (data.name && data.name !== data.ticker) ? data.name : '';
  document.title = data.ticker + ' — STX AI';
}

function renderIndicators(data) {
  // RSI
  if (data.rsi_14 != null) {
    const rsi = data.rsi_14;
    $('rsi-value').textContent = rsi.toFixed(1);
    $('rsi-zone').textContent = rsiZone(rsi);
    $('rsi-zone').className = 'card-sub ' + rsiZoneClass(rsi);
  } else {
    $('rsi-value').textContent = 'N/A';
  }

  // MACD
  if (data.macd != null) {
    $('macd-value').textContent = data.macd.toFixed(3);
    const parts = [];
    if (data.macd_signal != null) parts.push('Signal ' + data.macd_signal.toFixed(3));
    if (data.macd_histogram != null) {
      const h = data.macd_histogram;
      parts.push('Hist ' + (h >= 0 ? '+' : '') + h.toFixed(4));
    }
    $('macd-sub').textContent = parts.join(' · ');
  } else {
    $('macd-value').textContent = 'N/A';
  }

  // Moving Averages
  if (data.ma_50 != null) {
    $('ma50-value').textContent = '$' + data.ma_50.toLocaleString();
  } else {
    $('ma50-value').textContent = 'N/A';
  }
  if (data.ma_200 != null) {
    $('ma200-value').textContent = '$' + data.ma_200.toLocaleString();
  } else {
    $('ma200-value').textContent = 'N/A';
  }

  // MA relation
  if (data.price != null && data.ma_50 != null && data.ma_200 != null) {
    const above50 = data.price > data.ma_50;
    const above200 = data.price > data.ma_200;
    const pct50 = ((data.price / data.ma_50 - 1) * 100).toFixed(1);
    const pct200 = ((data.price / data.ma_200 - 1) * 100).toFixed(1);
    if (above50 && above200) {
      $('ma-relation').textContent = `Price ${pct50}% above MA50, ${pct200}% above MA200`;
      $('ma-relation').style.color = 'var(--green)';
    } else if (!above50 && !above200) {
      $('ma-relation').textContent = `Price ${pct50}% below MA50, ${pct200}% below MA200`;
      $('ma-relation').style.color = 'var(--red)';
    } else {
      $('ma-relation').textContent = `Mixed: ${above50 ? 'above' : 'below'} MA50, ${above200 ? 'above' : 'below'} MA200`;
      $('ma-relation').style.color = 'var(--yellow)';
    }
  }

  // Bollinger Bands
  if (data.bollinger_upper != null) {
    $('bb-upper').textContent = '$' + data.bollinger_upper.toLocaleString();
  }
  if (data.bollinger_middle != null) {
    $('bb-middle').textContent = '$' + data.bollinger_middle.toLocaleString();
  }
  if (data.bollinger_lower != null) {
    $('bb-lower').textContent = '$' + data.bollinger_lower.toLocaleString();
  }
}

function renderPatterns(data) {
  const container = $('patterns-list');

  // data might be { results: "..." } from scan endpoint or { patterns: [...] } from analyze
  let patterns = [];
  if (data.patterns && Array.isArray(data.patterns)) {
    patterns = data.patterns;
  }

  if (patterns.length === 0) {
    container.innerHTML = '<span class="muted">No significant patterns detected</span>';
    return;
  }

  container.innerHTML = patterns.map(p => {
    let cls = 'pattern-neutral';
    if (p.name.toLowerCase().includes('bull') || p.name.toLowerCase().includes('golden') ||
        p.name.toLowerCase().includes('oversold') || p.name.toLowerCase().includes('support') ||
        p.name.toLowerCase().includes('uptrend')) {
      cls = 'pattern-bullish';
    } else if (p.name.toLowerCase().includes('bear') || p.name.toLowerCase().includes('death') ||
               p.name.toLowerCase().includes('overbought') || p.name.toLowerCase().includes('resistance') ||
               p.name.toLowerCase().includes('downtrend')) {
      cls = 'pattern-bearish';
    }
    const icon = p.icon || '';
    return `<span class="pattern-tag ${cls}" title="${escHtml(p.action || '')}">${icon} ${escHtml(p.name)}</span>`;
  }).join('');
}

function renderNews(data) {
  const container = $('news-list');

  let articles = [];
  if (data.articles && Array.isArray(data.articles)) {
    // Each article may have { summary: "..." }
    articles = data.articles;
  }

  if (articles.length === 0) {
    container.innerHTML = '<span class="muted">No recent news available</span>';
    return;
  }

  // The news endpoint returns articles as LLM-generated summary text
  // Display as individual news items if possible, or as single block
  container.innerHTML = articles.map(a => {
    const text = a.summary || a.headline || '';
    // Try to parse bullet points
    const lines = text.split('\n').filter(l => l.trim().startsWith('-') || l.trim().startsWith('•'));
    if (lines.length > 0) {
      return lines.map(l => {
        const clean = l.replace(/^[-•]\s*/, '');
        return `<div class="news-item"><span class="news-headline">${escHtml(clean)}</span></div>`;
      }).join('');
    }
    return text ? `<div class="news-item">${escHtml(text)}</div>` : '';
  }).join('') || '<span class="muted">No recent news available</span>';
}

// ── Chat ──
chatForm.addEventListener('submit', async (e) => {
  e.preventDefault();
  const msg = chatInput.value.trim();
  if (!msg || !currentTicker) return;

  addChatMessage('user', msg);
  chatInput.value = '';

  const sendBtn = chatForm.querySelector('.chat-send');
  sendBtn.disabled = true;
  setStatus('AI thinking…', 'status-loading');

  try {
    const data = await fetchJSON('POST', '/api/v1/chat', {
      message: msg,
      ticker: currentTicker,
      lang: currentLang,
      deep_analysis: false,
    });
    const reply = data.reply || data.results || 'No response.';
    addChatMessage('ai', reply);
    setStatus('Ready for ' + currentTicker, '');
  } catch (e) {
    addChatMessage('ai', 'Error: ' + e.message);
    setStatus('Chat error: ' + e.message, 'status-error');
  } finally {
    sendBtn.disabled = false;
  }
});

function addChatMessage(role, text) {
  const div = document.createElement('div');
  div.className = 'chat-msg ' + role;
  div.textContent = text;
  chatMessages.appendChild(div);
  chatMessages.scrollTop = chatMessages.scrollHeight;
}

// ── Helpers ──
async function fetchJSON(method, url, body) {
  const opts = {
    method,
    headers: { 'Content-Type': 'application/json', 'Accept': 'application/json' },
  };
  if (body) opts.body = JSON.stringify(body);

  const resp = await fetch(url, opts);
  if (!resp.ok) {
    const text = await resp.text();
    throw new Error(resp.status + ': ' + (text || resp.statusText).slice(0, 200));
  }
  return resp.json();
}

function rsiZone(rsi) {
  if (rsi >= 70) return 'Overbought';
  if (rsi <= 30) return 'Oversold';
  return 'Neutral';
}

function rsiZoneClass(rsi) {
  if (rsi >= 70) return 'zone-overbought';
  if (rsi <= 30) return 'zone-oversold';
  return 'zone-neutral';
}

function detectMarket(ticker) {
  return /^\d+$/.test(ticker) ? 'hk' : 'us';
}

function escHtml(s) {
  if (!s) return '';
  const div = document.createElement('div');
  div.textContent = s;
  return div.innerHTML;
}

function setStatus(msg, cls) {
  statusText.textContent = msg;
  statusText.className = cls || '';
}

// ── Keyboard shortcuts ──
document.addEventListener('keydown', (e) => {
  // Cmd/Ctrl+K → focus ticker input
  if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
    e.preventDefault();
    tickerInput.focus();
    tickerInput.select();
  }
  // Escape → focus ticker
  if (e.key === 'Escape' && document.activeElement === chatInput) {
    tickerInput.focus();
    tickerInput.select();
  }
});
