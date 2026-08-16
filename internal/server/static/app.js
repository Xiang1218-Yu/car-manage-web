// 停车场管理系统 - 前端交互
(function () {
  'use strict';

  // 通用模态框：通过 data-modal 属性打开，data-close 关闭
  function openModal(id) {
    var el = document.getElementById(id);
    if (el) el.classList.add('open');
  }
  function closeModal(el) {
    if (!el) return;
    var modal = el.closest('.modal-backdrop');
    if (modal) modal.classList.remove('open');
  }
  document.addEventListener('click', function (e) {
    var opener = e.target.closest('[data-modal]');
    if (opener) {
      e.preventDefault();
      openModal(opener.getAttribute('data-modal'));
      // 自动填充隐藏字段并设置表单 action
      var fills = opener.getAttribute('data-fill');
      if (fills) {
        fills.split(',').forEach(function (pair) {
          var idx = pair.indexOf('=');
          if (idx < 0) return;
          var key = pair.slice(0, idx).trim();
          var val = pair.slice(idx + 1).trim();
          if (key === 'action') {
            // 设置同模态框内第一个 form 的 action
            var form = document.querySelector('#' + opener.getAttribute('data-modal') + ' form');
            if (form) form.setAttribute('action', val);
          } else {
            var input = document.querySelector('#' + opener.getAttribute('data-modal') + ' [name="' + key + '"]');
            if (input) input.value = val;
          }
        });
      }
      return;
    }
    if (e.target.closest('[data-close]') || e.target.classList.contains('modal-backdrop')) {
      closeModal(e.target);
    }
  });
  document.addEventListener('keydown', function (e) {
    if (e.key === 'Escape') {
      document.querySelectorAll('.modal-backdrop.open').forEach(function (m) { m.classList.remove('open'); });
    }
  });

  // 删除确认
  document.addEventListener('submit', function (e) {
    var form = e.target;
    if (form.getAttribute('data-confirm') && !confirm(form.getAttribute('data-confirm'))) {
      e.preventDefault();
    }
  });

  // 费用预估：监听规则/时间变化，调用 /api/fee/estimate
  var estimateBox = document.getElementById('fee-estimate');
  if (estimateBox) {
    var ruleSel = document.getElementById('est-rule');
    var inInput = document.getElementById('est-check-in');
    var outInput = document.getElementById('est-check-out');
    var refresh = function () {
      if (!ruleSel || !ruleSel.value) { return; }
      var params = new URLSearchParams();
      params.set('rule_id', ruleSel.value);
      if (inInput && inInput.value) {
        // 把 datetime-local 转为 RFC3339
        params.set('check_in', new Date(inInput.value).toISOString());
      } else {
        params.set('check_in', new Date(Date.now() - 3600 * 1000).toISOString());
      }
      if (outInput && outInput.value) {
        params.set('check_out', new Date(outInput.value).toISOString());
      }
      fetch('/api/fee/estimate?' + params.toString())
        .then(function (r) { return r.json(); })
        .then(function (bd) {
          if (bd.error) { estimateBox.querySelector('.total').textContent = '—'; return; }
          estimateBox.querySelector('.total').textContent = '¥' + bd.total_fee.toFixed(2);
          estimateBox.querySelector('.detail').innerHTML =
            '规则：' + escapeHtml(bd.rule_name) + '<br>' +
            '时长：' + bd.duration_minutes + ' 分钟<br>' +
            (bd.is_free ? '<b>免费</b>' :
              '首段：¥' + bd.first_block_fee.toFixed(2) +
              (bd.extra_units > 0 ? ' + ' + bd.extra_units + ' 单位 ¥' + bd.extra_block_fee.toFixed(2) : '') +
              (bd.daily_cap_applied ? '<br>（已按日封顶累计）' : ''));
        })
        .catch(function () { estimateBox.querySelector('.total').textContent = '—'; });
    };
    [ruleSel, inInput, outInput].forEach(function (el) {
      if (el) el.addEventListener('change', refresh);
      if (el) el.addEventListener('input', refresh);
    });
    refresh();
  }

  function escapeHtml(s) {
    return String(s).replace(/[&<>"']/g, function (c) {
      return { '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[c];
    });
  }
})();
