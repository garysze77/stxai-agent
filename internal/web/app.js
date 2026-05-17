// ── i18n ──
const T = {
  en: {
    welcome_title: 'Welcome to STX AI',
    welcome_subtitle: 'Enter a stock ticker above to see real-time data.',
    card_price: 'Price',
    card_rsi: 'RSI (14)',
    card_macd: 'MACD',
    card_ma: 'Moving Averages',
    card_ma50: 'MA50',
    card_ma200: 'MA200',
    card_patterns: 'Patterns',
    card_bb: 'Bollinger Bands',
    card_bb_upper: 'Upper',
    card_bb_middle: 'Middle',
    card_bb_lower: 'Lower',
    card_news: 'News',
    chat_placeholder: 'Ask anything about this stock…',
    chat_send: 'Send',
    status_ready: 'Ready — enter a ticker to begin',
    status_loading: 'Loading {ticker}…',
    status_price_loaded: 'Price loaded — fetching indicators…',
    status_ready_ticker: 'Dashboard ready for {ticker}',
    status_some_unavailable: 'Some data unavailable',
    status_ai_thinking: 'AI thinking…',
    status_price_failed: 'Failed to load price',
    status_chat_error: 'Chat error',
    rsi_overbought: 'Overbought',
    rsi_oversold: 'Oversold',
    rsi_neutral: 'Neutral',
    patterns_none: 'No significant patterns detected',
    patterns_scanning: 'Scanning…',
    news_none: 'No recent news available',
    news_loading: 'Loading news…',
    ma_above: 'Price {pct}% above MA{period}',
    ma_below: 'Price {pct}% below MA{period}',
    ma_mixed: 'Mixed: {rel50} MA50, {rel200} MA200',
    ma_above_label: 'above',
    ma_below_label: 'below',
  },
  'zh-HK': {
    welcome_title: '歡迎使用 STX AI',
    welcome_subtitle: '喺上面輸入股票代號查看即時數據。',
    card_price: '股價',
    card_rsi: 'RSI (14)',
    card_macd: 'MACD',
    card_ma: '移動平均線',
    card_ma50: 'MA50',
    card_ma200: 'MA200',
    card_patterns: '技術形態',
    card_bb: '保力加通道',
    card_bb_upper: '上軌',
    card_bb_middle: '中軌',
    card_bb_lower: '下軌',
    card_news: '新聞',
    chat_placeholder: '向 AI 查詢呢隻股票…',
    chat_send: '發送',
    status_ready: '就緒 — 輸入股票代號開始',
    status_loading: '載入 {ticker} 中…',
    status_price_loaded: '股價已載入 — 正在獲取技術指標…',
    status_ready_ticker: '{ticker} 儀表板已就緒',
    status_some_unavailable: '部分數據無法獲取',
    status_ai_thinking: 'AI 思考中…',
    status_price_failed: '無法載入股價',
    status_chat_error: '對話錯誤',
    rsi_overbought: '超買',
    rsi_oversold: '超賣',
    rsi_neutral: '中性',
    patterns_none: '未偵測到明顯技術形態',
    patterns_scanning: '掃描中…',
    news_none: '暫無相關新聞',
    news_loading: '載入新聞中…',
    ma_above: '股價高於 MA{period} {pct}%',
    ma_below: '股價低於 MA{period} {pct}%',
    ma_mixed: '混合：{rel50} MA50，{rel200} MA200',
    ma_above_label: '高於',
    ma_below_label: '低於',
  },
  'zh-CN': {
    welcome_title: '欢迎使用 STX AI',
    welcome_subtitle: '在上方输入股票代码查看实时数据。',
    card_price: '股价',
    card_rsi: 'RSI (14)',
    card_macd: 'MACD',
    card_ma: '移动平均线',
    card_ma50: 'MA50',
    card_ma200: 'MA200',
    card_patterns: '技术形态',
    card_bb: '布林带',
    card_bb_upper: '上轨',
    card_bb_middle: '中轨',
    card_bb_lower: '下轨',
    card_news: '新闻',
    chat_placeholder: '向 AI 查询该股票…',
    chat_send: '发送',
    status_ready: '就绪 — 输入股票代码开始',
    status_loading: '加载 {ticker} 中…',
    status_price_loaded: '股价已加载 — 正在获取技术指标…',
    status_ready_ticker: '{ticker} 仪表板已就绪',
    status_some_unavailable: '部分数据无法获取',
    status_ai_thinking: 'AI 思考中…',
    status_price_failed: '无法加载股价',
    status_chat_error: '对话错误',
    rsi_overbought: '超买',
    rsi_oversold: '超卖',
    rsi_neutral: '中性',
    patterns_none: '未检测到明显技术形态',
    patterns_scanning: '扫描中…',
    news_none: '暂无相关新闻',
    news_loading: '加载新闻中…',
    ma_above: '股价高于 MA{period} {pct}%',
    ma_below: '股价低于 MA{period} {pct}%',
    ma_mixed: '混合：{rel50} MA50，{rel200} MA200',
    ma_above_label: '高于',
    ma_below_label: '低于',
  },
};

function t(key, vars) {
  const dict = T[currentLang] || T['en'];
  let s = dict[key];
  if (s === undefined) {
    s = (T['en'][key]);
    if (s === undefined) return key;
  }
  if (vars) {
    for (const k of Object.keys(vars)) {
      s = s.replace('{' + k + '}', vars[k]);
    }
  }
  return s;
}

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
const chatSendBtn = chatForm ? chatForm.querySelector('.chat-send') : null;

// ── Language ──
function applyLang() {
  document.querySelectorAll('.lang-btn').forEach(btn => {
    btn.classList.toggle('active', btn.dataset.lang === currentLang);
  });
  document.querySelectorAll('[data-i18n]').forEach(el => {
    const key = el.dataset.i18n;
    if (key) el.textContent = t(key);
  });
  document.querySelectorAll('[data-i18n-placeholder]').forEach(el => {
    const key = el.dataset.i18nPlaceholder;
    if (key) el.placeholder = t(key);
  });
  if (currentTicker) {
    renderCardsForLang();
  }
  document.documentElement.lang = currentLang === 'zh-CN' ? 'zh-CN' : currentLang === 'zh-HK' ? 'zh-HK' : 'en';
}

document.querySelectorAll('.lang-btn').forEach(btn => {
  btn.addEventListener('click', () => {
    currentLang = btn.dataset.lang;
    applyLang();
  });
});

applyLang();

// ── Ticker search ──
tickerForm.addEventListener('submit', (e) => {
  e.preventDefault();
  const raw = tickerInput.value.trim();
  if (!raw) return;
  currentTicker = raw.toUpperCase();
  loadDashboard(currentTicker);
});

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
  setStatus(t('status_loading', { ticker: ticker }), 'status-loading');

  resetCards();

  // 1. Price first (fast)
  try {
    const priceData = await fetchJSON('GET', '/api/v1/price/' + ticker);
    renderPrice(priceData);
    setStatus(t('status_price_loaded'), 'status-loading');
  } catch (e) {
    setStatus(t('status_price_failed') + ': ' + e.message, 'status-error');
    return;
  }

  // 2. Indicators (includes patterns) + News in parallel
  const results = await Promise.allSettled([
    fetchJSON('GET', '/api/v1/indicators/' + ticker),
    fetchJSON('GET', '/api/v1/news/' + ticker + '?lang=' + currentLang),
  ]);

  const [indResult, newsResult] = results;

  if (indResult.status === 'fulfilled') {
    renderIndicators(indResult.value);
    renderPatterns(indResult.value);
  }
  if (newsResult.status === 'fulfilled') {
    renderNews(newsResult.value);
  }

  const errors = [indResult, newsResult]
    .filter(r => r.status === 'rejected')
    .map(r => r.reason.message);

  if (errors.length > 0) {
    setStatus(t('status_some_unavailable') + ': ' + errors.join('; '), 'status-error');
  } else {
    setStatus(t('status_ready_ticker', { ticker: ticker }), '');
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
  $('patterns-list').innerHTML = '<span class="muted">' + t('patterns_scanning') + '</span>';
  $('bb-upper').textContent = '…';
  $('bb-middle').textContent = '…';
  $('bb-lower').textContent = '…';
  $('news-list').innerHTML = '<span class="muted">' + t('news_loading') + '</span>';
}

// Re-render card labels + RSI zone + patterns/placeholders on lang switch
function renderCardsForLang() {
  // Re-render RSI zone label
  const rsiEl = $('rsi-value');
  const rsiText = rsiEl.textContent;
  if (rsiText && rsiText !== '…' && rsiText !== 'N/A') {
    const rsi = parseFloat(rsiText);
    if (!isNaN(rsi)) {
      $('rsi-zone').textContent = t(rsiZoneKey(rsi));
    }
  }

  // Re-render MA relation
  const maRel = $('ma-relation');
  if (maRel.dataset.maState) {
    try {
      const state = JSON.parse(maRel.dataset.maState);
      renderMARelation(state.price, state.ma50, state.ma200);
    } catch(e) {}
  }

  // Re-render patterns
  const patternsList = $('patterns-list');
  if (patternsList.dataset.patternsJson) {
    try {
      const patterns = JSON.parse(patternsList.dataset.patternsJson);
      renderPatternsFromData(patterns);
    } catch(e) {}
  }
}

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
    $('rsi-zone').textContent = t(rsiZoneKey(rsi));
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
    renderMARelation(data.price, data.ma_50, data.ma_200);
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

function renderMARelation(price, ma50, ma200) {
  const above50 = price > ma50;
  const above200 = price > ma200;
  const pct50 = ((price / ma50 - 1) * 100).toFixed(1);
  const pct200 = ((price / ma200 - 1) * 100).toFixed(1);

  const state = { price: price, ma50: ma50, ma200: ma200 };
  $('ma-relation').dataset.maState = JSON.stringify(state);

  if (above50 && above200) {
    $('ma-relation').textContent =
      t('ma_above', { pct: pct50, period: '50' }) + ', ' +
      t('ma_above', { pct: pct200, period: '200' });
    $('ma-relation').style.color = 'var(--green)';
  } else if (!above50 && !above200) {
    $('ma-relation').textContent =
      t('ma_below', { pct: pct50, period: '50' }) + ', ' +
      t('ma_below', { pct: pct200, period: '200' });
    $('ma-relation').style.color = 'var(--red)';
  } else {
    const rel50 = above50 ? t('ma_above_label') : t('ma_below_label');
    const rel200 = above200 ? t('ma_above_label') : t('ma_below_label');
    $('ma-relation').textContent = t('ma_mixed', { rel50: rel50, rel200: rel200 });
    $('ma-relation').style.color = 'var(--yellow)';
  }
}

function renderPatterns(data) {
  let patterns = [];
  if (data.patterns && Array.isArray(data.patterns)) {
    patterns = data.patterns;
  }
  $('patterns-list').dataset.patternsJson = JSON.stringify(patterns);
  renderPatternsFromData(patterns);
}

function renderPatternsFromData(patterns) {
  const container = $('patterns-list');
  if (patterns.length === 0) {
    container.innerHTML = '<span class="muted">' + t('patterns_none') + '</span>';
    return;
  }
  container.innerHTML = patterns.map(p => {
    let cls = 'pattern-neutral';
    const nameLower = p.name.toLowerCase();
    if (/bull|golden|oversold|support|uptrend/.test(nameLower)) {
      cls = 'pattern-bullish';
    } else if (/bear|death|overbought|resistance|downtrend/.test(nameLower)) {
      cls = 'pattern-bearish';
    }
    const icon = p.icon || '';
    return '<span class="pattern-tag ' + cls + '" title="' + escHtml(p.action || '') + '">' + icon + ' ' + escHtml(p.name) + '</span>';
  }).join('');
}

function renderNews(data) {
  const container = $('news-list');
  let articles = [];
  if (data.articles && Array.isArray(data.articles)) {
    articles = data.articles;
  }

  if (articles.length === 0) {
    container.innerHTML = '<span class="muted">' + t('news_none') + '</span>';
    return;
  }

  container.innerHTML = articles.map(a => {
    const text = a.summary || a.headline || '';
    const lines = text.split('\n').filter(l => l.trim().startsWith('-') || l.trim().startsWith('•'));
    if (lines.length > 0) {
      return lines.map(l => {
        const clean = l.replace(/^[-•]\s*/, '');
        return '<div class="news-item"><span class="news-headline">' + escHtml(clean) + '</span></div>';
      }).join('');
    }
    return text ? '<div class="news-item">' + escHtml(text) + '</div>' : '';
  }).join('') || '<span class="muted">' + t('news_none') + '</span>';
}

// ── Chat ──
chatForm.addEventListener('submit', async (e) => {
  e.preventDefault();
  const msg = chatInput.value.trim();
  if (!msg || !currentTicker) return;

  addChatMessage('user', msg);
  chatInput.value = '';

  if (chatSendBtn) chatSendBtn.disabled = true;
  setStatus(t('status_ai_thinking'), 'status-loading');

  try {
    const data = await fetchJSON('POST', '/api/v1/chat', {
      message: msg,
      ticker: currentTicker,
      lang: currentLang,
      deep_analysis: false,
    });
    const reply = data.reply || data.results || 'No response.';
    addChatMessage('ai', reply);
    setStatus(t('status_ready_ticker', { ticker: currentTicker }), '');
  } catch (e) {
    addChatMessage('ai', 'Error: ' + e.message);
    setStatus(t('status_chat_error') + ': ' + e.message, 'status-error');
  } finally {
    if (chatSendBtn) chatSendBtn.disabled = false;
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

function rsiZoneKey(rsi) {
  if (rsi >= 70) return 'rsi_overbought';
  if (rsi <= 30) return 'rsi_oversold';
  return 'rsi_neutral';
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
  if ((e.metaKey || e.ctrlKey) && e.key === 'k') {
    e.preventDefault();
    tickerInput.focus();
    tickerInput.select();
  }
  if (e.key === 'Escape' && document.activeElement === chatInput) {
    tickerInput.focus();
    tickerInput.select();
  }
});
