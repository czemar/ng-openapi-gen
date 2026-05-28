(function () {
  'use strict';

  const DEFAULT_SPEC = `openapi: "3.0.0"
info:
  title: Petstore
  version: "1.0"
paths:
  /pets:
    get:
      operationId: listPets
      parameters:
        - name: limit
          in: query
          schema:
            type: integer
      responses:
        "200":
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/Pets"
    post:
      operationId: createPet
      responses:
        "201":
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/Pet"
  /pets/{petId}:
    get:
      operationId: showPetById
      parameters:
        - name: petId
          in: path
          required: true
          schema:
            type: string
      responses:
        "200":
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/Pet"
components:
  schemas:
    Pet:
      type: object
      required: [id, name]
      properties:
        id:
          type: integer
        name:
          type: string
        tag:
          type: string
    Pets:
      type: array
      items:
        $ref: "#/components/schemas/Pet"
    Error:
      type: object
      required: [code, message]
      properties:
        code:
          type: integer
        message:
          type: string
`;

  let wasmReady = false;
  let wasmQueue = [];

  function loadWasm() {
    if (!window.WebAssembly) {
      document.getElementById('demo-app').innerHTML =
        '<p class="hx:text-red-500">WebAssembly is not supported in this browser.</p>';
      return;
    }
    const go = new Go();
    WebAssembly.instantiateStreaming(fetch('demo.wasm'), go.importObject)
      .then(function (result) {
        go.run(result.instance);
        wasmReady = true;
        document.getElementById('generate-btn').disabled = false;
        document.getElementById('status').textContent = 'Ready';
        for (const fn of wasmQueue) fn();
        wasmQueue = [];
      })
      .catch(function (err) {
        document.getElementById('status').textContent = 'Failed to load WASM: ' + err.message;
      });
  }

  function renderApp() {
    const app = document.getElementById('demo-app');
    app.innerHTML =
      '<div style="display:flex;gap:1rem;min-height:500px;flex-wrap:wrap">' +
      '  <div style="flex:1;min-width:300px;display:flex;flex-direction:column;gap:0.5rem">' +
      '    <label style="font-weight:600;font-size:0.875rem">OpenAPI Spec (YAML/JSON)</label>' +
      '    <textarea id="spec-input" spellcheck="false" style="flex:1;min-height:400px;font-family:ui-monospace,monospace;font-size:0.8rem;padding:0.75rem;border:1px solid #d1d5db;border-radius:6px;resize:vertical;tab-size:2;background:#fafafa">' +
      DEFAULT_SPEC.replace(/[&<>]/g, function (c) {
        return { '&': '&amp;', '<': '&lt;', '>': '&gt;' }[c];
      }) +
      '</textarea>' +
      '    <div style="display:flex;gap:0.5rem;align-items:center">' +
      '      <button id="generate-btn" disabled style="padding:0.5rem 1.5rem;font-weight:600;border:none;border-radius:6px;cursor:pointer;background:#2563eb;color:white;font-size:0.875rem">Generate</button>' +
      '      <span id="status" style="font-size:0.8rem;color:#6b7280">Loading WASM\u2026</span>' +
      '    </div>' +
      '  </div>' +
      '  <div id="output-panel" style="flex:1;min-width:300px;display:flex;flex-direction:column;gap:0.5rem">' +
      '    <label style="font-weight:600;font-size:0.875rem">Generated Files</label>' +
      '    <div id="file-tree" style="font-family:ui-monospace,monospace;font-size:0.8rem;border:1px solid #d1d5db;border-radius:6px;padding:0.5rem;background:#fafafa;max-height:200px;overflow-y:auto;display:none"></div>' +
      '    <pre id="file-content" style="flex:1;min-height:350px;font-family:ui-monospace,monospace;font-size:0.78rem;border:1px solid #d1d5db;border-radius:6px;padding:0.75rem;overflow:auto;background:#1e1e1e;color:#d4d4d4;margin:0;white-space:pre;tab-size:2">' +
      '      <span style="color:#6b7280">Generated files will appear here\u2026</span>' +
      '    </pre>' +
      '  </div>' +
      '</div>';

    document.getElementById('generate-btn').addEventListener('click', generate);
  }

  function generate() {
    const btn = document.getElementById('generate-btn');
    const status = document.getElementById('status');
    const spec = document.getElementById('spec-input').value;

    if (!spec.trim()) {
      status.textContent = 'Please enter an OpenAPI spec';
      return;
    }

    btn.disabled = true;
    status.textContent = 'Generating\u2026';

    function doGenerate() {
      var startTime = performance.now();
      try {
        var resultJson = window.generateOpenAPIDemo(spec, '{}');
        var result = JSON.parse(resultJson);
        var elapsed = (performance.now() - startTime).toFixed(0);

        if (result.error) {
          status.textContent = 'Error: ' + result.error;
          btn.disabled = false;
          return;
        }

        status.textContent = 'Generated ' + Object.keys(result).length + ' files in ' + elapsed + 'ms';
        renderFiles(result);
      } catch (e) {
        status.textContent = 'Error: ' + e.message;
      }
      btn.disabled = false;
    }

    if (wasmReady) {
      doGenerate();
    } else {
      wasmQueue.push(doGenerate);
      status.textContent = 'Waiting for WASM\u2026';
    }
  }

  function renderFiles(files) {
    var treeEl = document.getElementById('file-tree');
    var contentEl = document.getElementById('file-content');
    var paths = Object.keys(files).sort();

    treeEl.style.display = 'block';
    treeEl.innerHTML = buildFileTree(paths);
    contentEl.textContent = 'Select a file to view';

    treeEl.querySelectorAll('[data-file]').forEach(function (el) {
      el.addEventListener('click', function (e) {
        var filePath = e.currentTarget.dataset.file;
        treeEl.querySelectorAll('[data-file]').forEach(function (x) {
          x.style.fontWeight = 'normal';
          x.style.color = '';
        });
        e.currentTarget.style.fontWeight = '600';
        e.currentTarget.style.color = '#2563eb';
        contentEl.textContent = files[filePath];
      });
    });

    // Auto-select first file
    var first = treeEl.querySelector('[data-file]');
    if (first) {
      first.style.fontWeight = '600';
      first.style.color = '#2563eb';
      contentEl.textContent = files[first.dataset.file];
    }
  }

  function buildFileTree(paths) {
    var result = '';
    var dirs = {};
    for (var i = 0; i < paths.length; i++) {
      var parts = paths[i].split('/');
      if (parts.length > 1) {
        var dir = parts[0];
        if (!dirs[dir]) dirs[dir] = [];
        dirs[dir].push(parts.slice(1).join('/'));
      } else {
        result += '<div data-file="' + escapeHtml(paths[i]) + '" style="cursor:pointer;padding:2px 4px;border-radius:3px;display:flex;align-items:center;gap:4px">' +
          '  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="#2563eb" stroke-width="2"><path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z"/><polyline points="14 2 14 8 20 8"/></svg>' +
          '  ' + escapeHtml(paths[i]) +
          '</div>';
      }
    }
    var dirNames = Object.keys(dirs).sort();
    for (var j = 0; j < dirNames.length; j++) {
      var dn = dirNames[j];
      result += '<div style="margin-top:4px">' +
        '  <div style="cursor:default;padding:2px 4px;display:flex;align-items:center;gap:4px;font-weight:600;font-size:0.75rem;color:#6b7280;text-transform:uppercase">' +
        '    <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="#6b7280" stroke-width="2"><path d="M22 19a2 2 0 01-2 2H4a2 2 0 01-2-2V5a2 2 0 012-2h5l2 3h9a2 2 0 012 2z"/></svg>' +
        '    ' + escapeHtml(dn) +
        '  </div>';
      var children = dirs[dn].sort();
      for (var k = 0; k < children.length; k++) {
        var fullPath = dn + '/' + children[k];
        result += '<div data-file="' + escapeHtml(fullPath) + '" style="cursor:pointer;padding:2px 4px 2px 24px;border-radius:3px;display:flex;align-items:center;gap:4px">' +
          '  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="#2563eb" stroke-width="2"><path d="M14 2H6a2 2 0 00-2 2v16a2 2 0 002 2h12a2 2 0 002-2V8z"/><polyline points="14 2 14 8 20 8"/></svg>' +
          '  ' + escapeHtml(children[k]) +
          '</div>';
      }
      result += '</div>';
    }
    return result;
  }

  function escapeHtml(s) {
    return s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;');
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', function () { renderApp(); loadWasm(); });
  } else {
    renderApp();
    loadWasm();
  }
})();
