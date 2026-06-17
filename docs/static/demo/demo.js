import { createHighlighter } from 'https://esm.sh/shiki@1.27.2';

const DEFAULT_SPEC = 'openapi: "3.0.0"\ninfo:\n  title: Petstore\n  version: "1.0"\npaths:\n  /pets:\n    get:\n      operationId: listPets\n      parameters:\n        - name: limit\n          in: query\n          schema:\n            type: integer\n      responses:\n        "200":\n          content:\n            application/json:\n              schema:\n                $ref: "#/components/schemas/Pets"\n    post:\n      operationId: createPet\n      responses:\n        "201":\n          content:\n            application/json:\n              schema:\n                $ref: "#/components/schemas/Pet"\n  /pets/{petId}:\n    get:\n      operationId: showPetById\n      parameters:\n        - name: petId\n          in: path\n          required: true\n          schema:\n            type: string\n      responses:\n        "200":\n          content:\n            application/json:\n              schema:\n                $ref: "#/components/schemas/Pet"\ncomponents:\n  schemas:\n    Pet:\n      type: object\n      required: [id, name]\n      properties:\n        id:\n          type: integer\n        name:\n          type: string\n        tag:\n          type: string\n    Pets:\n      type: array\n      items:\n        $ref: "#/components/schemas/Pet"\n    Error:\n      type: object\n      required: [code, message]\n      properties:\n        code:\n          type: integer\n        message:\n          type: string\n';

const DEFAULT_CONFIG = {
  promises: true,
  services: false,
  enumStyle: 'alias',
  modelSuffix: '',
  serviceSuffix: 'Service',
  camelizeModelNames: true,
  ignoreUnusedModels: true,
  skipJsonSuffix: false,
};

const highlighterPromise = createHighlighter({
  themes: ['github-dark'],
  langs: ['typescript', 'yaml', 'json', 'plaintext'],
});

const STYLES = [
  '.demo-tabs{display:flex;gap:0;border-bottom:1px solid #d1d5db;margin-top:1rem}',
  '.dark .demo-tabs{border-color:#374151}',
  '.demo-tab{padding:.5rem 1.25rem;font-size:.875rem;font-weight:500;cursor:pointer;color:#6b7280;border-bottom:2px solid transparent;transition:all .15s;user-select:none}',
  '.dark .demo-tab{color:#9ca3af}',
  '.demo-tab:hover{color:#374151}',
  '.dark .demo-tab:hover{color:#d1d5db}',
  '.demo-tab.active{color:#2563eb;border-bottom-color:#2563eb;font-weight:600}',
  '.dark .demo-tab.active{color:#60a5fa;border-bottom-color:#60a5fa}',
  '.demo-panel{display:none}',
  '.demo-panel.active{display:flex;flex-direction:column;gap:.75rem;min-height:450px}',
  '.demo-label{font-size:.75rem;font-weight:600;text-transform:uppercase;letter-spacing:.05em;color:#6b7280}',
  '.dark .demo-label{color:#9ca3af}',
  '.demo-textarea{width:100%;min-height:350px;flex:1;font-family:ui-monospace,SFMono-Regular,Menlo,Monaco,Consolas,monospace;font-size:.8125rem;line-height:1.5;padding:.75rem;border:1px solid #d1d5db;border-radius:6px;resize:vertical;tab-size:2;background:#f9fafb;color:#1f2937;box-sizing:border-box}',
  '.dark .demo-textarea{background:#1e1e1e;color:#d4d4d4;border-color:#374151}',
  '.demo-textarea:focus{outline:2px solid #2563eb;outline-offset:-1px;border-color:transparent}',
  '.demo-toolbar{display:flex;gap:.5rem;align-items:center;flex-wrap:wrap}',
  '.demo-btn{display:inline-flex;align-items:center;gap:.375rem;padding:.5rem 1.25rem;font-size:.875rem;font-weight:600;border:none;border-radius:6px;cursor:pointer;color:#fff;background:#2563eb;transition:background .15s}',
  '.demo-btn:hover{background:#1d4ed8}',
  '.demo-btn:disabled{background:#93c5fd;cursor:not-allowed}',
  '.demo-btn-secondary{background:#6b7280}',
  '.demo-btn-secondary:hover{background:#4b5563}',
  '.demo-status{font-size:.8125rem;color:#6b7280}',
  '.dark .demo-status{color:#9ca3af}',
  '.demo-status.error{color:#ef4444}',
  '.dark .demo-status.error{color:#f87171}',
  '.demo-preview-split{display:flex;gap:1rem;flex:1;min-height:0}',
  '.demo-preview-tree{width:220px;flex-shrink:0;border:1px solid #d1d5db;border-radius:6px;padding:.375rem;background:#f9fafb;max-height:500px;overflow-y:auto;font-family:ui-monospace,SFMono-Regular,Menlo,Monaco,Consolas,monospace;font-size:.8125rem}',
  '.dark .demo-preview-tree{background:#1e1e1e;border-color:#374151}',
  '.demo-preview-code{flex:1;min-width:0;border:1px solid #d1d5db;border-radius:6px;overflow:hidden;background:#1e1e1e;display:flex;flex-direction:column}',
  '.dark .demo-preview-code{border-color:#374151}',
  '.demo-preview-code .demo-code-filename{flex-shrink:0}',
  '.demo-preview-code .shiki{flex:1;overflow:auto;padding:.75rem;margin:0;font-size:.75rem;line-height:1.6}',
  '.demo-preview-code pre.shiki{flex:1;overflow:auto;padding:.75rem;margin:0;font-size:.8125rem;line-height:1.5;background:transparent!important}',
  '.demo-preview-code pre.shiki code{background:transparent!important;border:none;counter-reset:none;font-size:inherit;line-height:inherit}',
  '.demo-preview-code pre.shiki .line{display:block;padding:0;margin:0;height:auto;line-height:0;background:transparent!important;border:none}',
  '.demo-preview-code pre:not(.shiki){flex:1;overflow:auto;padding:.75rem;margin:0;font-size:.75rem;line-height:1.6;color:#d4d4d4;font-family:ui-monospace,SFMono-Regular,Menlo,Monaco,Consolas,monospace}',
  '.demo-fileitem{display:flex;align-items:center;gap:.375rem;padding:.25rem .5rem;border-radius:4px;cursor:pointer;color:#374151;transition:background .1s}',
  '.dark .demo-fileitem{color:#d1d5db}',
  '.demo-fileitem:hover{background:#e5e7eb}',
  '.dark .demo-fileitem:hover{background:#374151}',
  '.demo-fileitem.active{font-weight:600;color:#2563eb;background:#eff6ff}',
  '.dark .demo-fileitem.active{color:#60a5fa;background:#1e3a5f}',
  '.demo-dirheader{display:flex;align-items:center;gap:.375rem;padding:.25rem .5rem;font-size:.6875rem;font-weight:700;text-transform:uppercase;letter-spacing:.05em;color:#9ca3af;margin-top:.25rem}',
  '.dark .demo-dirheader{color:#6b7280}',
  '.demo-dirchildren{padding-left:1.25rem}',
  '.demo-placeholder{display:flex;align-items:center;justify-content:center;flex:1;color:#6b7280;font-style:italic;font-size:.875rem}',
  '.dark .demo-placeholder{color:#9ca3af}',
  '.demo-error-box{padding:.75rem;background:#fef2f2;border:1px solid #fecaca;border-radius:6px;color:#dc2626;font-size:.875rem}',
  '.dark .demo-error-box{background:#450a0a;border-color:#7f1d1d;color:#fca5a5}',
  '.demo-loading{display:inline-block;width:1rem;height:1rem;border:2px solid #d1d5db;border-top-color:#2563eb;border-radius:50%;animation:spin .6s linear infinite}',
  '@keyframes spin{to{transform:rotate(360deg)}}',
  '.demo-spinner{display:inline-block;width:1rem;height:1rem;border:2px solid #e5e7eb;border-top-color:#2563eb;border-radius:50%;animation:spin .6s linear infinite;vertical-align:middle;margin-right:.375rem}',
  '.demo-options{border:1px solid #d1d5db;border-radius:6px;overflow:hidden}',
  '.dark .demo-options{border-color:#374151}',
  '.demo-options-header{display:flex;align-items:center;justify-content:space-between;padding:.5rem .75rem;cursor:pointer;font-size:.8125rem;font-weight:600;color:#374151;background:#f3f4f6;user-select:none}',
  '.dark .demo-options-header{color:#d1d5db;background:#262626}',
  '.demo-options-header:hover{background:#e5e7eb}',
  '.dark .demo-options-header:hover{background:#333}',
  '.demo-options-body{padding:.5rem .75rem .75rem;display:none}',
  '.demo-options-body.open{display:block}',
  '.demo-opt-group{display:flex;flex-wrap:wrap;gap:.75rem;margin-bottom:.5rem}',
  '.demo-opt-group:last-child{margin-bottom:0}',
  '.demo-opt{display:flex;flex-direction:column;gap:.125rem;min-width:140px;flex:1}',
  '.demo-opt-label{font-size:.75rem;font-weight:500;color:#374151}',
  '.dark .demo-opt-label{color:#d1d5db}',
  '.demo-opt-desc{font-size:.6875rem;color:#9ca3af;line-height:1.3}',
  '.demo-opt-toggle{display:inline-flex;align-items:center;gap:.5rem;cursor:pointer;padding:.25rem 0}',
  '.demo-opt-toggle input{width:2.25rem;height:1.25rem;appearance:none;background:#d1d5db;border-radius:999px;position:relative;cursor:pointer;transition:background .15s;flex-shrink:0}',
  '.demo-opt-toggle input::before{content:"";position:absolute;top:2px;left:2px;width:calc(1.25rem - 4px);height:calc(1.25rem - 4px);background:#fff;border-radius:50%;transition:transform .15s}',
  '.demo-opt-toggle input:checked{background:#2563eb}',
  '.demo-opt-toggle input:checked::before{transform:translateX(1rem)}',
  '.demo-opt-select{padding:.375rem .5rem;font-size:.75rem;border:1px solid #d1d5db;border-radius:4px;background:#fff;color:#374151}',
  '.dark .demo-opt-select{background:#1e1e1e;color:#d1d5db;border-color:#374151}',
  '.demo-opt-input{padding:.375rem .5rem;font-size:.75rem;border:1px solid #d1d5db;border-radius:4px;background:#fff;color:#374151;width:100%;box-sizing:border-box}',
  '.dark .demo-opt-input{background:#1e1e1e;color:#d1d5db;border-color:#374151}',
  '.demo-opt-radio-group{display:flex;gap:.5rem;padding:.125rem 0}',
  '.demo-opt-radio{cursor:pointer;display:flex;align-items:center;gap:.25rem;font-size:.75rem;color:#374151}',
  '.dark .demo-opt-radio{color:#d1d5db}',
  '.demo-opt-radio input{accent-color:#2563eb}',
  '.demo-fileitem-empty{padding:1rem;text-align:center;color:#9ca3af;font-style:italic;font-size:.8125rem}',
].join('');

const styleTag = document.createElement('style');
styleTag.textContent = STYLES;
document.head.appendChild(styleTag);

let wasmReady = false;
let wasmQueue = [];
let currentFiles = null;

function escapeHtml(s) {
  return s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;');
}

function buildOptionsHTML() {
  const checked = (key) => DEFAULT_CONFIG[key] ? 'checked' : '';
  const selected = (key, val) => DEFAULT_CONFIG[key] === val ? 'selected' : '';

  return `
    <div class="demo-options-header" id="opt-toggle">
      <span>Generation Options</span>
      <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" id="opt-chevron"><polyline points="9 18 15 12 9 6"/></svg>
    </div>
    <div class="demo-options-body" id="opt-body">
      <div class="demo-opt-group">
        <div class="demo-opt">
          <span class="demo-opt-label">Response style</span>
          <div class="demo-opt-radio-group">
            <label class="demo-opt-radio"><input type="radio" name="response-style" value="promise" ${checked('promises') ? 'checked' : ''}> Promise&lt;T&gt;</label>
            <label class="demo-opt-radio"><input type="radio" name="response-style" value="observable"> Observable&lt;T&gt;</label>
          </div>
          <span class="demo-opt-desc">Whether generated functions return Promises or RxJS Observables</span>
        </div>
        <div class="demo-opt">
          <span class="demo-opt-label">Services</span>
          <label class="demo-opt-toggle">
            <input type="checkbox" id="opt-services" ${checked('services')}>
            Generate
          </label>
          <span class="demo-opt-desc">Generate Angular @Injectable service classes per tag group</span>
        </div>
        <div class="demo-opt">
          <span class="demo-opt-label">Enum style</span>
          <select class="demo-opt-select" id="opt-enum-style">
            <option value="alias" ${selected('enumStyle', 'alias')}>alias — Type union</option>
            <option value="upper" ${selected('enumStyle', 'upper')}>upper — UPPER_CASE enum</option>
            <option value="ignorecase" ${selected('enumStyle', 'ignorecase')}>ignorecase — Original case enum</option>
          </select>
          <span class="demo-opt-desc">How enum types are generated in TypeScript</span>
        </div>
      </div>
      <div class="demo-opt-group">
        <div class="demo-opt">
          <span class="demo-opt-label">Model suffix</span>
          <input class="demo-opt-input" id="opt-model-suffix" type="text" value="${escapeHtml(DEFAULT_CONFIG.modelSuffix)}" placeholder="e.g. DTO">
          <span class="demo-opt-desc">Suffix appended to all model class names</span>
        </div>
        <div class="demo-opt">
          <span class="demo-opt-label">Service suffix</span>
          <input class="demo-opt-input" id="opt-service-suffix" type="text" value="${escapeHtml(DEFAULT_CONFIG.serviceSuffix)}" placeholder="e.g. Service">
          <span class="demo-opt-desc">Suffix appended to all service class names</span>
        </div>
      </div>
      <div class="demo-opt-group">
        <div class="demo-opt">
          <span class="demo-opt-label">Camelize model names</span>
          <label class="demo-opt-toggle">
            <input type="checkbox" id="opt-camelize" ${checked('camelizeModelNames')}>
            On
          </label>
          <span class="demo-opt-desc">Convert model names to camelCase (e.g. MyModel → myModel)</span>
        </div>
        <div class="demo-opt">
          <span class="demo-opt-label">Ignore unused models</span>
          <label class="demo-opt-toggle">
            <input type="checkbox" id="opt-ignore-unused" ${checked('ignoreUnusedModels')}>
            On
          </label>
          <span class="demo-opt-desc">Remove models not referenced by any operation</span>
        </div>
        <div class="demo-opt">
          <span class="demo-opt-label">Skip $Json suffix</span>
          <label class="demo-opt-toggle">
            <input type="checkbox" id="opt-skip-json" ${checked('skipJsonSuffix')}>
            On
          </label>
          <span class="demo-opt-desc">Omit "$Json" suffix when only JSON response exists</span>
        </div>
      </div>
    </div>`;
}

function buildConfig() {
  const config = {};
  const responseStyle = document.querySelector('input[name="response-style"]:checked');
  if (responseStyle) config.promises = responseStyle.value === 'promise';
  config.services = document.getElementById('opt-services').checked;
  config.enumStyle = document.getElementById('opt-enum-style').value;
  const modelSuffix = document.getElementById('opt-model-suffix').value.trim();
  if (modelSuffix) config.modelSuffix = modelSuffix;
  const serviceSuffix = document.getElementById('opt-service-suffix').value.trim();
  if (serviceSuffix && serviceSuffix !== 'Service') config.serviceSuffix = serviceSuffix;
  config.camelizeModelNames = document.getElementById('opt-camelize').checked;
  config.ignoreUnusedModels = document.getElementById('opt-ignore-unused').checked;
  config.skipJsonSuffix = document.getElementById('opt-skip-json').checked;
  return JSON.stringify(config);
}

function resetOptions() {
  const radios = document.querySelectorAll('input[name="response-style"]');
  radios.forEach((r) => { r.checked = r.value === 'promise'; });
  document.getElementById('opt-services').checked = DEFAULT_CONFIG.services;
  document.getElementById('opt-enum-style').value = DEFAULT_CONFIG.enumStyle;
  document.getElementById('opt-model-suffix').value = DEFAULT_CONFIG.modelSuffix;
  document.getElementById('opt-service-suffix').value = DEFAULT_CONFIG.serviceSuffix;
  document.getElementById('opt-camelize').checked = DEFAULT_CONFIG.camelizeModelNames;
  document.getElementById('opt-ignore-unused').checked = DEFAULT_CONFIG.ignoreUnusedModels;
  document.getElementById('opt-skip-json').checked = DEFAULT_CONFIG.skipJsonSuffix;
}

function switchTab(name) {
  document.querySelectorAll('.demo-tab').forEach((t) => t.classList.toggle('active', t.dataset.tab === name));
  document.querySelectorAll('.demo-panel').forEach((p) => p.classList.toggle('active', p.id === 'panel-' + name));
}

function renderApp() {
  const app = document.getElementById('demo-app');
  app.innerHTML =
    '<div class="demo-tabs">' +
    '  <div class="demo-tab active" data-tab="spec" id="tab-spec">Spec</div>' +
    '  <div class="demo-tab" data-tab="preview" id="tab-preview">Preview</div>' +
    '</div>' +
    '<div class="demo-panel active" id="panel-spec">' +
    '  <textarea id="spec-input" class="demo-textarea" spellcheck="false" placeholder="Paste your OpenAPI 3.0/3.1 spec here\u2026"></textarea>' +
    '  <div class="demo-options">' + buildOptionsHTML() + '</div>' +
    '  <div class="demo-toolbar">' +
    '    <button id="generate-btn" class="demo-btn" disabled><svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><polygon points="5 3 19 12 5 21 5 3"/></svg> Generate</button>' +
    '    <button id="reset-btn" class="demo-btn demo-btn-secondary">Reset</button>' +
    '    <span id="status" class="demo-status"><span class="demo-loading"></span> Loading WASM\u2026</span>' +
    '  </div>' +
    '</div>' +
    '<div class="demo-panel" id="panel-preview">' +
    '  <div class="demo-preview-split">' +
    '    <div id="preview-tree" class="demo-preview-tree"><div class="demo-fileitem-empty">No files yet — generate from the Spec tab</div></div>' +
    '    <div id="preview-code" class="demo-preview-code"><div class="demo-placeholder">Generated code will appear here</div></div>' +
    '  </div>' +
    '</div>';

  document.getElementById('spec-input').value = DEFAULT_SPEC;

  document.getElementById('tab-spec').addEventListener('click', () => switchTab('spec'));
  document.getElementById('tab-preview').addEventListener('click', () => switchTab('preview'));

  document.getElementById('spec-input').addEventListener('keydown', (e) => {
    if ((e.metaKey || e.ctrlKey) && e.key === 'Enter') generate();
  });

  document.getElementById('opt-toggle').addEventListener('click', () => {
    document.getElementById('opt-body').classList.toggle('open');
    const chevron = document.getElementById('opt-chevron');
    chevron.style.transform = document.getElementById('opt-body').classList.contains('open') ? 'rotate(90deg)' : '';
    chevron.style.transition = 'transform .15s';
  });

  document.getElementById('generate-btn').addEventListener('click', generate);
  document.getElementById('reset-btn').addEventListener('click', () => {
    document.getElementById('spec-input').value = DEFAULT_SPEC;
    document.getElementById('preview-tree').innerHTML = '<div class="demo-fileitem-empty">No files yet — generate from the Spec tab</div>';
    document.getElementById('preview-code').innerHTML = '<div class="demo-placeholder">Generated code will appear here</div>';
    currentFiles = null;
    resetOptions();
    switchTab('spec');
    document.getElementById('status').textContent = 'Reset to default spec and options';
    document.getElementById('status').className = 'demo-status';
  });
}

function loadWasm() {
  if (!window.WebAssembly) {
    document.getElementById('status').textContent = 'WebAssembly not supported in this browser';
    document.getElementById('status').className = 'demo-status error';
    return;
  }
  if (typeof Go === 'undefined') {
    setTimeout(loadWasm, 200);
    return;
  }
  const go = new Go();
  WebAssembly.instantiateStreaming(fetch('demo.wasm'), go.importObject)
    .then((result) => {
      go.run(result.instance);
      wasmReady = true;
      document.getElementById('generate-btn').disabled = false;
      document.getElementById('status').textContent = 'Ready \u2014 paste a spec, adjust options, click Generate';
      document.getElementById('status').className = 'demo-status';
      wasmQueue.forEach((fn) => fn());
      wasmQueue = [];
    })
    .catch((err) => {
      document.getElementById('status').textContent = 'Failed to load WASM: ' + err.message;
      document.getElementById('status').className = 'demo-status error';
    });
}

function generate() {
  const btn = document.getElementById('generate-btn');
  const status = document.getElementById('status');
  const spec = document.getElementById('spec-input').value;

  if (!spec.trim()) {
    status.textContent = 'Please enter an OpenAPI spec';
    status.className = 'demo-status error';
    return;
  }

  btn.disabled = true;
  status.innerHTML = '<span class="demo-spinner"></span> Generating\u2026';
  status.className = 'demo-status';

  function run() {
    try {
      const config = buildConfig();
      const start = performance.now();
      const resultJson = window.generateOpenAPIDemo(spec, config);
      const result = JSON.parse(resultJson);
      const elapsed = (performance.now() - start).toFixed(0);

      if (result.error) {
        status.textContent = 'Error: ' + result.error;
        status.className = 'demo-status error';
        btn.disabled = false;
        return;
      }

      const keys = Object.keys(result);
      if (keys.length === 0) {
        status.textContent = 'No files generated (empty spec?)';
        status.className = 'demo-status error';
        btn.disabled = false;
        return;
      }

      status.textContent = keys.length + ' file' + (keys.length === 1 ? '' : 's') + ' generated in ' + elapsed + 'ms';
      status.className = 'demo-status';
      currentFiles = result;
      renderFiles(result);
      switchTab('preview');
    } catch (e) {
      status.textContent = 'Error: ' + e.message;
      status.className = 'demo-status error';
    }
    btn.disabled = false;
  }

  if (wasmReady) { run(); } else { wasmQueue.push(run); status.innerHTML = '<span class="demo-spinner"></span> Waiting for WASM\u2026'; }
}

function renderFiles(files) {
  const paths = Object.keys(files).sort();
  const treeEl = document.getElementById('preview-tree');
  const codeEl = document.getElementById('preview-code');

  treeEl.innerHTML = buildFileTree(paths);

  const items = treeEl.querySelectorAll('[data-file]');
  items.forEach((el) => {
    el.addEventListener('click', () => {
      items.forEach((x) => x.classList.remove('active'));
      el.classList.add('active');
      showFile(el.dataset.file, files[el.dataset.file]);
    });
  });

  if (items.length > 0) {
    items[0].classList.add('active');
    showFile(items[0].dataset.file, files[items[0].dataset.file]);
  }
}

async function showFile(path, content) {
  const codeEl = document.getElementById('preview-code');
  const ext = path.split('.').pop();
  const lang = ext === 'ts' ? 'typescript' : ext === 'yaml' || ext === 'yml' ? 'yaml' : ext === 'json' ? 'json' : 'plaintext';

  const header = '<div class="demo-code-filename" style="padding:.375rem .75rem;font-size:.75rem;color:#9ca3af;background:#252526;border-bottom:1px solid #333;font-family:ui-monospace,monospace;flex-shrink:0">' + escapeHtml(path) + '</div>';

  try {
    const highlighter = await highlighterPromise;
    const highlighted = highlighter.codeToHtml(content, { lang, theme: 'github-dark' });
    codeEl.innerHTML = header + highlighted;
  } catch {
    codeEl.innerHTML = header + '<pre>' + escapeHtml(content) + '</pre>';
  }
}

function buildFileTree(paths) {
  const dirs = {};
  let html = '';

  for (let i = 0; i < paths.length; i++) {
    const parts = paths[i].split('/');
    if (parts.length > 1) {
      const dir = parts[0];
      if (!dirs[dir]) dirs[dir] = [];
      dirs[dir].push(parts.slice(1).join('/'));
    }
  }

  const rootFiles = paths.filter((p) => !p.includes('/')).sort();
  rootFiles.forEach((f) => {
    html += '<div class="demo-fileitem" data-file="' + escapeHtml(f) + '">' + fileIcon() + ' ' + escapeHtml(f) + '</div>';
  });

  Object.keys(dirs).sort().forEach((d) => {
    html += '<div><div class="demo-dirheader">' + folderIcon() + ' ' + escapeHtml(d) + '</div><div class="demo-dirchildren">';
    dirs[d].sort().forEach((f) => {
      html += '<div class="demo-fileitem" data-file="' + escapeHtml(d + '/' + f) + '">' + fileIcon() + ' ' + escapeHtml(f) + '</div>';
    });
    html += '</div></div>';
  });

  return html;
}

function fileIcon() {
  return '<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="#2563eb" stroke-width="2" style="flex-shrink:0"><path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z"/><polyline points="14 2 14 8 20 8"/></svg>';
}

function folderIcon() {
  return '<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="#9ca3af" stroke-width="2" style="flex-shrink:0"><path d="M22 19a2 2 0 01-2 2H4a2 2 0 01-2-2V5a2 2 0 012-2h5l2 3h9a2 2 0 012 2z"/></svg>';
}

if (document.readyState === 'loading') {
  document.addEventListener('DOMContentLoaded', () => { renderApp(); setTimeout(loadWasm, 100); });
} else {
  renderApp();
  setTimeout(loadWasm, 100);
}
