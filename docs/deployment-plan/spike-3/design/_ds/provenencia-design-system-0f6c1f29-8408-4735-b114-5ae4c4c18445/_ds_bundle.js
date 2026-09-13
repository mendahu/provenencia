/* @ds-bundle: {"format":4,"namespace":"ProvenenciaDesignSystem_0f6c1f","components":[{"name":"Badge","sourcePath":"components/core/Badge.jsx"},{"name":"Button","sourcePath":"components/core/Button.jsx"},{"name":"Card","sourcePath":"components/core/Card.jsx"},{"name":"Icon","sourcePath":"components/core/Icon.jsx"},{"name":"IconButton","sourcePath":"components/core/IconButton.jsx"},{"name":"Tag","sourcePath":"components/core/Tag.jsx"},{"name":"Tooltip","sourcePath":"components/core/Tooltip.jsx"},{"name":"Callout","sourcePath":"components/feedback/Callout.jsx"},{"name":"Dialog","sourcePath":"components/feedback/Dialog.jsx"},{"name":"EmptyState","sourcePath":"components/feedback/EmptyState.jsx"},{"name":"Toast","sourcePath":"components/feedback/Toast.jsx"},{"name":"Checkbox","sourcePath":"components/forms/Checkbox.jsx"},{"name":"Field","sourcePath":"components/forms/Field.jsx"},{"name":"Input","sourcePath":"components/forms/Input.jsx"},{"name":"Radio","sourcePath":"components/forms/Radio.jsx"},{"name":"Select","sourcePath":"components/forms/Select.jsx"},{"name":"Switch","sourcePath":"components/forms/Switch.jsx"},{"name":"Breadcrumbs","sourcePath":"components/navigation/Breadcrumbs.jsx"},{"name":"SidebarNav","sourcePath":"components/navigation/SidebarNav.jsx"},{"name":"Tabs","sourcePath":"components/navigation/Tabs.jsx"},{"name":"EvidenceBadge","sourcePath":"components/research/EvidenceBadge.jsx"},{"name":"FactRow","sourcePath":"components/research/FactRow.jsx"},{"name":"PersonChip","sourcePath":"components/research/PersonChip.jsx"},{"name":"SourceCitation","sourcePath":"components/research/SourceCitation.jsx"}],"sourceHashes":{"components/core/Badge.jsx":"e022318ccd05","components/core/Button.jsx":"21253f784ffb","components/core/Card.jsx":"3db918f63aea","components/core/Icon.jsx":"4b7aab438a56","components/core/IconButton.jsx":"5513e43c56b1","components/core/Tag.jsx":"24480e5ea018","components/core/Tooltip.jsx":"c40d75bc0e2c","components/feedback/Callout.jsx":"38fa4898554a","components/feedback/Dialog.jsx":"5c1950409fc0","components/feedback/EmptyState.jsx":"29c059dabcf1","components/feedback/Toast.jsx":"e121a42c3d78","components/forms/Checkbox.jsx":"ece343f50b46","components/forms/Field.jsx":"856e666d1f65","components/forms/Input.jsx":"cc5ce9a458ae","components/forms/Radio.jsx":"a95d04945366","components/forms/Select.jsx":"cb9afaaceabe","components/forms/Switch.jsx":"621b2cbb0136","components/navigation/Breadcrumbs.jsx":"49fe11b3bd33","components/navigation/SidebarNav.jsx":"9b980f4f6648","components/navigation/Tabs.jsx":"671c34cea982","components/research/EvidenceBadge.jsx":"e76cca9d9073","components/research/FactRow.jsx":"c4f562d561a3","components/research/PersonChip.jsx":"f55b189574e3","components/research/SourceCitation.jsx":"6523e139628e","ui_kits/app/App.jsx":"7720149073fd","ui_kits/app/Dashboard.jsx":"887b930f8bb5","ui_kits/app/PersonView.jsx":"e2c0885562c7","ui_kits/app/Shell.jsx":"f951fec6bb20","ui_kits/app/SourcesLibrary.jsx":"2b21ec175744","ui_kits/app/TreeView.jsx":"c1280883c8f4","ui_kits/site/Site.jsx":"a0d9fadb5985"},"inlinedExternals":[],"unexposedExports":[]} */

(() => {

const __ds_ns = (window.ProvenenciaDesignSystem_0f6c1f = window.ProvenenciaDesignSystem_0f6c1f || {});

const __ds_scope = {};

(__ds_ns.__errors = __ds_ns.__errors || []);

// components/core/Card.jsx
try { (() => {
function _extends() { return _extends = Object.assign ? Object.assign.bind() : function (n) { for (var e = 1; e < arguments.length; e++) { var t = arguments[e]; for (var r in t) ({}).hasOwnProperty.call(t, r) && (n[r] = t[r]); } return n; }, _extends.apply(null, arguments); }
function Card({
  children,
  title,
  subtitle,
  actions,
  footer,
  padding = 'var(--space-8)',
  tone = 'card',
  hoverable,
  style,
  ...rest
}) {
  const [hover, setHover] = React.useState(false);
  const bg = tone === 'sunken' ? 'var(--surface-sunken)' : tone === 'raised' ? 'var(--surface-raised)' : 'var(--surface-card)';
  return /*#__PURE__*/React.createElement("section", _extends({
    onMouseEnter: () => setHover(true),
    onMouseLeave: () => setHover(false),
    style: {
      background: bg,
      border: '1px solid var(--border-subtle)',
      borderRadius: 'var(--radius-md)',
      boxShadow: hoverable && hover ? 'var(--shadow-md)' : 'var(--shadow-sm)',
      transform: hoverable && hover ? 'translateY(var(--lift-hover))' : 'none',
      transition: 'box-shadow var(--dur-normal) var(--ease-standard), transform var(--dur-normal) var(--ease-standard)',
      overflow: 'hidden',
      ...style
    }
  }, rest), (title || actions) && /*#__PURE__*/React.createElement("header", {
    style: {
      display: 'flex',
      alignItems: 'baseline',
      justifyContent: 'space-between',
      gap: 'var(--space-6)',
      padding: `var(--space-6) ${padding} var(--space-5)`,
      borderBottom: '1px solid var(--border-subtle)'
    }
  }, /*#__PURE__*/React.createElement("div", null, title && /*#__PURE__*/React.createElement("h3", {
    style: {
      fontFamily: 'var(--font-display)',
      fontSize: 'var(--text-h3)',
      fontWeight: 'var(--weight-semibold)',
      color: 'var(--text-display)',
      margin: 0
    }
  }, title), subtitle && /*#__PURE__*/React.createElement("p", {
    style: {
      margin: '3px 0 0',
      fontSize: 'var(--text-caption)',
      color: 'var(--text-muted)'
    }
  }, subtitle)), actions), /*#__PURE__*/React.createElement("div", {
    style: {
      padding
    }
  }, children), footer && /*#__PURE__*/React.createElement("footer", {
    style: {
      padding: `var(--space-5) ${padding}`,
      borderTop: '1px solid var(--border-subtle)',
      background: 'var(--surface-sunken)',
      fontSize: 'var(--text-caption)',
      color: 'var(--text-muted)'
    }
  }, footer));
}
Object.assign(__ds_scope, { Card });
})(); } catch (e) { __ds_ns.__errors.push({ path: "components/core/Card.jsx", error: String((e && e.message) || e) }); }

// components/core/Icon.jsx
try { (() => {
function _extends() { return _extends = Object.assign ? Object.assign.bind() : function (n) { for (var e = 1; e < arguments.length; e++) { var t = arguments[e]; for (var r in t) ({}).hasOwnProperty.call(t, r) && (n[r] = t[r]); } return n; }, _extends.apply(null, arguments); }
const CDN = 'https://unpkg.com/lucide-static@0.469.0/icons/';
function Icon({
  name,
  size = 16,
  strokeWeight,
  style,
  ...rest
}) {
  const url = `url(${CDN}${name}.svg)`;
  return /*#__PURE__*/React.createElement("i", _extends({
    "aria-hidden": "true",
    "data-icon": name,
    style: {
      display: 'inline-block',
      flex: 'none',
      width: size,
      height: size,
      backgroundColor: 'currentColor',
      verticalAlign: '-0.125em',
      WebkitMaskImage: url,
      maskImage: url,
      WebkitMaskRepeat: 'no-repeat',
      maskRepeat: 'no-repeat',
      WebkitMaskPosition: 'center',
      maskPosition: 'center',
      WebkitMaskSize: 'contain',
      maskSize: 'contain',
      opacity: strokeWeight === 'light' ? 0.7 : 1,
      ...style
    }
  }, rest));
}
Object.assign(__ds_scope, { Icon });
})(); } catch (e) { __ds_ns.__errors.push({ path: "components/core/Icon.jsx", error: String((e && e.message) || e) }); }

// components/core/Badge.jsx
try { (() => {
function _extends() { return _extends = Object.assign ? Object.assign.bind() : function (n) { for (var e = 1; e < arguments.length; e++) { var t = arguments[e]; for (var r in t) ({}).hasOwnProperty.call(t, r) && (n[r] = t[r]); } return n; }, _extends.apply(null, arguments); }
const tones = {
  neutral: ['var(--surface-inset)', 'var(--text-secondary)', 'var(--border-default)'],
  accent: ['var(--accent-soft)', 'var(--accent-soft-fg)', 'var(--accent-line)'],
  success: ['var(--success-soft)', 'var(--success-fg)', 'var(--success)'],
  warning: ['var(--warning-soft)', 'var(--warning-fg)', 'var(--warning)'],
  danger: ['var(--danger-soft)', 'var(--danger-fg)', 'var(--danger)'],
  info: ['var(--info-soft)', 'var(--info-fg)', 'var(--info)']
};
function Badge({
  children,
  tone = 'neutral',
  icon,
  subtle,
  style,
  ...rest
}) {
  const [bg, fg, line] = tones[tone] || tones.neutral;
  return /*#__PURE__*/React.createElement("span", _extends({
    style: {
      display: 'inline-flex',
      alignItems: 'center',
      gap: 5,
      padding: '2px 8px',
      borderRadius: 'var(--radius-xs)',
      background: subtle ? 'transparent' : bg,
      color: fg,
      border: `1px solid ${subtle ? line : 'transparent'}`,
      fontFamily: 'var(--font-body)',
      fontSize: 'var(--text-micro)',
      fontWeight: 'var(--weight-semibold)',
      letterSpacing: 'var(--tracking-caps)',
      textTransform: 'uppercase',
      lineHeight: 1.6,
      whiteSpace: 'nowrap',
      ...style
    }
  }, rest), icon ? /*#__PURE__*/React.createElement(__ds_scope.Icon, {
    name: icon,
    size: 11
  }) : null, children);
}
Object.assign(__ds_scope, { Badge });
})(); } catch (e) { __ds_ns.__errors.push({ path: "components/core/Badge.jsx", error: String((e && e.message) || e) }); }

// components/core/Button.jsx
try { (() => {
function _extends() { return _extends = Object.assign ? Object.assign.bind() : function (n) { for (var e = 1; e < arguments.length; e++) { var t = arguments[e]; for (var r in t) ({}).hasOwnProperty.call(t, r) && (n[r] = t[r]); } return n; }, _extends.apply(null, arguments); }
const sizes = {
  sm: {
    height: 'var(--control-h-sm)',
    padding: '0 10px',
    font: 'var(--text-caption)',
    gap: 6,
    icon: 13
  },
  md: {
    height: 'var(--control-h-md)',
    padding: '0 14px',
    font: 'var(--text-body-sm)',
    gap: 7,
    icon: 15
  },
  lg: {
    height: 'var(--control-h-lg)',
    padding: '0 20px',
    font: 'var(--text-body)',
    gap: 9,
    icon: 17
  }
};
const variants = {
  primary: {
    background: 'var(--accent)',
    color: 'var(--accent-fg)',
    border: '1px solid var(--accent)'
  },
  secondary: {
    background: 'var(--surface-raised)',
    color: 'var(--text-primary)',
    border: '1px solid var(--border-default)'
  },
  ghost: {
    background: 'transparent',
    color: 'var(--text-secondary)',
    border: '1px solid transparent'
  },
  danger: {
    background: 'var(--danger)',
    color: 'var(--paper-0)',
    border: '1px solid var(--danger)'
  },
  link: {
    background: 'transparent',
    color: 'var(--text-link)',
    border: '1px solid transparent',
    padding: 0,
    height: 'auto',
    textDecoration: 'underline',
    textUnderlineOffset: 2
  }
};
function Button({
  children,
  variant = 'secondary',
  size = 'md',
  iconLeft,
  iconRight,
  disabled,
  loading,
  fullWidth,
  style,
  ...rest
}) {
  const s = sizes[size] || sizes.md;
  const v = variants[variant] || variants.secondary;
  const [hover, setHover] = React.useState(false);
  const [down, setDown] = React.useState(false);
  const hoverStyle = disabled || variant === 'link' ? null : {
    primary: {
      background: 'var(--accent-hover)',
      borderColor: 'var(--accent-hover)'
    },
    secondary: {
      background: 'var(--surface-sunken)',
      borderColor: 'var(--border-strong)'
    },
    ghost: {
      background: 'var(--surface-hover)',
      color: 'var(--text-primary)'
    },
    danger: {
      background: 'var(--madder-700)',
      borderColor: 'var(--madder-700)'
    }
  }[variant];
  return /*#__PURE__*/React.createElement("button", _extends({
    disabled: disabled || loading,
    onMouseEnter: () => setHover(true),
    onMouseLeave: () => {
      setHover(false);
      setDown(false);
    },
    onMouseDown: () => setDown(true),
    onMouseUp: () => setDown(false),
    style: {
      display: fullWidth ? 'flex' : 'inline-flex',
      width: fullWidth ? '100%' : undefined,
      alignItems: 'center',
      justifyContent: 'center',
      gap: s.gap,
      height: s.height,
      padding: s.padding,
      fontFamily: 'var(--font-body)',
      fontSize: s.font,
      fontWeight: 'var(--weight-medium)',
      letterSpacing: 'var(--tracking-wide)',
      borderRadius: 'var(--radius-sm)',
      cursor: disabled ? 'not-allowed' : 'pointer',
      transition: 'background var(--dur-fast) var(--ease-standard), color var(--dur-fast) var(--ease-standard), border-color var(--dur-fast) var(--ease-standard), transform var(--dur-instant) var(--ease-standard)',
      opacity: disabled ? 0.45 : 1,
      whiteSpace: 'nowrap',
      transform: down ? 'scale(var(--press-scale))' : 'none',
      ...v,
      ...(hover ? hoverStyle : null),
      ...style
    }
  }, rest), loading ? /*#__PURE__*/React.createElement(__ds_scope.Icon, {
    name: "loader",
    size: s.icon
  }) : iconLeft ? /*#__PURE__*/React.createElement(__ds_scope.Icon, {
    name: iconLeft,
    size: s.icon
  }) : null, children, iconRight ? /*#__PURE__*/React.createElement(__ds_scope.Icon, {
    name: iconRight,
    size: s.icon
  }) : null);
}
Object.assign(__ds_scope, { Button });
})(); } catch (e) { __ds_ns.__errors.push({ path: "components/core/Button.jsx", error: String((e && e.message) || e) }); }

// components/core/IconButton.jsx
try { (() => {
function _extends() { return _extends = Object.assign ? Object.assign.bind() : function (n) { for (var e = 1; e < arguments.length; e++) { var t = arguments[e]; for (var r in t) ({}).hasOwnProperty.call(t, r) && (n[r] = t[r]); } return n; }, _extends.apply(null, arguments); }
const dims = {
  sm: 26,
  md: 32,
  lg: 40
};
function IconButton({
  icon,
  label,
  size = 'md',
  variant = 'ghost',
  active,
  disabled,
  style,
  ...rest
}) {
  const d = dims[size] || dims.md;
  const [hover, setHover] = React.useState(false);
  const bg = active ? 'var(--surface-selected)' : hover && !disabled ? 'var(--surface-hover)' : 'transparent';
  return /*#__PURE__*/React.createElement("button", _extends({
    "aria-label": label,
    title: label,
    disabled: disabled,
    onMouseEnter: () => setHover(true),
    onMouseLeave: () => setHover(false),
    style: {
      width: d,
      height: d,
      display: 'inline-flex',
      alignItems: 'center',
      justifyContent: 'center',
      borderRadius: 'var(--radius-sm)',
      cursor: disabled ? 'not-allowed' : 'pointer',
      background: variant === 'outline' ? 'var(--surface-raised)' : bg,
      border: variant === 'outline' ? '1px solid var(--border-default)' : '1px solid transparent',
      color: active ? 'var(--accent)' : 'var(--text-secondary)',
      opacity: disabled ? 0.4 : 1,
      transition: 'background var(--dur-fast) var(--ease-standard), color var(--dur-fast) var(--ease-standard)',
      ...style
    }
  }, rest), /*#__PURE__*/React.createElement(__ds_scope.Icon, {
    name: icon,
    size: size === 'sm' ? 14 : size === 'lg' ? 19 : 16
  }));
}
Object.assign(__ds_scope, { IconButton });
})(); } catch (e) { __ds_ns.__errors.push({ path: "components/core/IconButton.jsx", error: String((e && e.message) || e) }); }

// components/core/Tag.jsx
try { (() => {
function _extends() { return _extends = Object.assign ? Object.assign.bind() : function (n) { for (var e = 1; e < arguments.length; e++) { var t = arguments[e]; for (var r in t) ({}).hasOwnProperty.call(t, r) && (n[r] = t[r]); } return n; }, _extends.apply(null, arguments); }
function Tag({
  children,
  color = 'var(--paper-500)',
  onRemove,
  interactive,
  style,
  ...rest
}) {
  const [hover, setHover] = React.useState(false);
  return /*#__PURE__*/React.createElement("span", _extends({
    onMouseEnter: () => setHover(true),
    onMouseLeave: () => setHover(false),
    style: {
      display: 'inline-flex',
      alignItems: 'center',
      gap: 6,
      padding: '3px 9px 3px 8px',
      borderRadius: 'var(--radius-pill)',
      background: 'var(--surface-raised)',
      border: '1px solid var(--border-subtle)',
      color: 'var(--text-secondary)',
      fontFamily: 'var(--font-body)',
      fontSize: 'var(--text-caption)',
      boxShadow: hover && interactive ? 'var(--shadow-sm)' : 'none',
      cursor: interactive ? 'pointer' : 'default',
      transition: 'box-shadow var(--dur-fast) var(--ease-standard)',
      ...style
    }
  }, rest), /*#__PURE__*/React.createElement("span", {
    style: {
      width: 7,
      height: 7,
      borderRadius: '50%',
      background: color,
      flex: 'none'
    }
  }), children, onRemove ? /*#__PURE__*/React.createElement("button", {
    onClick: onRemove,
    "aria-label": "Remove",
    style: {
      display: 'inline-flex',
      border: 0,
      background: 'transparent',
      padding: 0,
      marginLeft: 1,
      color: 'var(--text-faint)',
      cursor: 'pointer'
    }
  }, /*#__PURE__*/React.createElement(__ds_scope.Icon, {
    name: "x",
    size: 12
  })) : null);
}
Object.assign(__ds_scope, { Tag });
})(); } catch (e) { __ds_ns.__errors.push({ path: "components/core/Tag.jsx", error: String((e && e.message) || e) }); }

// components/core/Tooltip.jsx
try { (() => {
function _extends() { return _extends = Object.assign ? Object.assign.bind() : function (n) { for (var e = 1; e < arguments.length; e++) { var t = arguments[e]; for (var r in t) ({}).hasOwnProperty.call(t, r) && (n[r] = t[r]); } return n; }, _extends.apply(null, arguments); }
function Tooltip({
  children,
  label,
  side = 'top',
  style,
  ...rest
}) {
  const [open, setOpen] = React.useState(false);
  const pos = {
    top: {
      bottom: '100%',
      left: '50%',
      transform: 'translate(-50%,-6px)'
    },
    bottom: {
      top: '100%',
      left: '50%',
      transform: 'translate(-50%,6px)'
    },
    left: {
      right: '100%',
      top: '50%',
      transform: 'translate(-6px,-50%)'
    },
    right: {
      left: '100%',
      top: '50%',
      transform: 'translate(6px,-50%)'
    }
  }[side];
  return /*#__PURE__*/React.createElement("span", _extends({
    onMouseEnter: () => setOpen(true),
    onMouseLeave: () => setOpen(false),
    onFocus: () => setOpen(true),
    onBlur: () => setOpen(false),
    style: {
      position: 'relative',
      display: 'inline-flex',
      ...style
    }
  }, rest), children, /*#__PURE__*/React.createElement("span", {
    role: "tooltip",
    style: {
      position: 'absolute',
      ...pos,
      zIndex: 40,
      pointerEvents: 'none',
      opacity: open ? 1 : 0,
      transition: 'opacity var(--dur-fast) var(--ease-standard)',
      background: 'var(--paper-900)',
      color: 'var(--paper-50)',
      fontFamily: 'var(--font-body)',
      fontSize: 'var(--text-micro)',
      letterSpacing: 'var(--tracking-wide)',
      padding: '4px 8px',
      borderRadius: 'var(--radius-xs)',
      whiteSpace: 'nowrap',
      boxShadow: 'var(--shadow-md)'
    }
  }, label));
}
Object.assign(__ds_scope, { Tooltip });
})(); } catch (e) { __ds_ns.__errors.push({ path: "components/core/Tooltip.jsx", error: String((e && e.message) || e) }); }

// components/feedback/Callout.jsx
try { (() => {
function _extends() { return _extends = Object.assign ? Object.assign.bind() : function (n) { for (var e = 1; e < arguments.length; e++) { var t = arguments[e]; for (var r in t) ({}).hasOwnProperty.call(t, r) && (n[r] = t[r]); } return n; }, _extends.apply(null, arguments); }
const tones = {
  info: {
    color: 'var(--info)',
    soft: 'var(--info-soft)',
    fg: 'var(--info-fg)',
    icon: 'info'
  },
  success: {
    color: 'var(--success)',
    soft: 'var(--success-soft)',
    fg: 'var(--success-fg)',
    icon: 'check-check'
  },
  warning: {
    color: 'var(--warning)',
    soft: 'var(--warning-soft)',
    fg: 'var(--warning-fg)',
    icon: 'triangle-alert'
  },
  danger: {
    color: 'var(--danger)',
    soft: 'var(--danger-soft)',
    fg: 'var(--danger-fg)',
    icon: 'octagon-alert'
  },
  neutral: {
    color: 'var(--border-strong)',
    soft: 'var(--surface-sunken)',
    fg: 'var(--text-secondary)',
    icon: 'info'
  }
};
function Callout({
  tone = 'info',
  title,
  children,
  icon,
  actions,
  onDismiss,
  variant = 'soft',
  compact,
  detail,
  style,
  ...rest
}) {
  const t = tones[tone] || tones.info;
  const plain = variant === 'plain';
  return /*#__PURE__*/React.createElement("div", _extends({
    role: tone === 'danger' ? 'alert' : 'status',
    style: {
      display: 'flex',
      alignItems: 'flex-start',
      gap: compact ? 'var(--space-5)' : 'var(--space-6)',
      padding: compact ? 'var(--space-5) var(--space-6)' : 'var(--space-6) var(--space-7)',
      background: plain ? 'transparent' : t.soft,
      border: `1px solid ${plain ? 'var(--border-subtle)' : 'transparent'}`,
      borderLeft: `3px solid ${t.color}`,
      borderRadius: 'var(--radius-sm)',
      color: plain ? 'var(--text-primary)' : t.fg,
      ...style
    }
  }, rest), /*#__PURE__*/React.createElement("span", {
    style: {
      color: t.color,
      display: 'flex',
      marginTop: title ? 3 : 2,
      flex: 'none'
    }
  }, /*#__PURE__*/React.createElement(__ds_scope.Icon, {
    name: icon || t.icon,
    size: compact ? 14 : 15
  })), /*#__PURE__*/React.createElement("div", {
    style: {
      flex: 1,
      minWidth: 0,
      display: 'flex',
      flexDirection: 'column',
      gap: 3
    }
  }, title && /*#__PURE__*/React.createElement("div", {
    style: {
      fontFamily: 'var(--font-display)',
      fontSize: compact ? 'var(--text-h4)' : 'var(--text-h3)',
      fontWeight: 'var(--weight-semibold)',
      color: plain ? 'var(--text-display)' : t.fg,
      lineHeight: 'var(--lh-snug)'
    }
  }, title), children && /*#__PURE__*/React.createElement("div", {
    style: {
      fontSize: compact ? 'var(--text-caption)' : 'var(--text-body-sm)',
      lineHeight: 'var(--lh-relaxed)',
      color: plain ? 'var(--text-secondary)' : t.fg,
      opacity: plain ? 1 : 0.92,
      maxWidth: 'var(--measure-prose)'
    }
  }, children), detail && /*#__PURE__*/React.createElement("pre", {
    style: {
      margin: 'var(--space-4) 0 0',
      padding: 'var(--space-5)',
      background: 'var(--surface-card)',
      border: '1px solid var(--border-subtle)',
      borderRadius: 'var(--radius-xs)',
      fontFamily: 'var(--font-mono)',
      fontSize: 'var(--text-micro)',
      lineHeight: 1.65,
      color: 'var(--text-secondary)',
      overflowX: 'auto',
      whiteSpace: 'pre-wrap',
      wordBreak: 'break-word'
    }
  }, detail), actions && /*#__PURE__*/React.createElement("div", {
    style: {
      display: 'flex',
      gap: 'var(--space-5)',
      alignItems: 'center',
      marginTop: 'var(--space-5)'
    }
  }, actions)), onDismiss && /*#__PURE__*/React.createElement(__ds_scope.IconButton, {
    icon: "x",
    label: "Dismiss",
    size: "sm",
    onClick: onDismiss
  }));
}
Object.assign(__ds_scope, { Callout });
})(); } catch (e) { __ds_ns.__errors.push({ path: "components/feedback/Callout.jsx", error: String((e && e.message) || e) }); }

// components/feedback/Dialog.jsx
try { (() => {
function _extends() { return _extends = Object.assign ? Object.assign.bind() : function (n) { for (var e = 1; e < arguments.length; e++) { var t = arguments[e]; for (var r in t) ({}).hasOwnProperty.call(t, r) && (n[r] = t[r]); } return n; }, _extends.apply(null, arguments); }
function Dialog({
  open = true,
  title,
  subtitle,
  children,
  footer,
  width = 520,
  onClose,
  style,
  ...rest
}) {
  if (!open) return null;
  return /*#__PURE__*/React.createElement("div", {
    style: {
      position: 'absolute',
      inset: 0,
      display: 'grid',
      placeItems: 'center',
      background: 'var(--surface-overlay)',
      backdropFilter: 'blur(2px)',
      zIndex: 60,
      padding: 'var(--space-9)'
    }
  }, /*#__PURE__*/React.createElement("div", _extends({
    role: "dialog",
    "aria-modal": "true",
    style: {
      width,
      maxWidth: '100%',
      background: 'var(--surface-card)',
      border: '1px solid var(--border-default)',
      borderRadius: 'var(--radius-md)',
      boxShadow: 'var(--shadow-overlay)',
      overflow: 'hidden',
      animation: 'none',
      ...style
    }
  }, rest), /*#__PURE__*/React.createElement("header", {
    style: {
      display: 'flex',
      alignItems: 'flex-start',
      gap: 'var(--space-6)',
      padding: 'var(--space-7) var(--space-8) var(--space-6)',
      borderBottom: '1px solid var(--border-subtle)'
    }
  }, /*#__PURE__*/React.createElement("div", {
    style: {
      flex: 1
    }
  }, /*#__PURE__*/React.createElement("h2", {
    style: {
      margin: 0,
      fontFamily: 'var(--font-display)',
      fontSize: 'var(--text-h2)',
      fontWeight: 'var(--weight-semibold)',
      color: 'var(--text-display)',
      letterSpacing: 'var(--tracking-display)'
    }
  }, title), subtitle && /*#__PURE__*/React.createElement("p", {
    style: {
      margin: '4px 0 0',
      fontSize: 'var(--text-body-sm)',
      color: 'var(--text-muted)'
    }
  }, subtitle)), onClose && /*#__PURE__*/React.createElement(__ds_scope.IconButton, {
    icon: "x",
    label: "Close",
    onClick: onClose
  })), /*#__PURE__*/React.createElement("div", {
    style: {
      padding: 'var(--space-8)',
      fontSize: 'var(--text-body-sm)',
      color: 'var(--text-secondary)',
      lineHeight: 'var(--lh-relaxed)'
    }
  }, children), footer && /*#__PURE__*/React.createElement("footer", {
    style: {
      display: 'flex',
      justifyContent: 'flex-end',
      gap: 'var(--space-5)',
      padding: 'var(--space-6) var(--space-8)',
      borderTop: '1px solid var(--border-subtle)',
      background: 'var(--surface-sunken)'
    }
  }, footer)));
}
Object.assign(__ds_scope, { Dialog });
})(); } catch (e) { __ds_ns.__errors.push({ path: "components/feedback/Dialog.jsx", error: String((e && e.message) || e) }); }

// components/feedback/EmptyState.jsx
try { (() => {
function _extends() { return _extends = Object.assign ? Object.assign.bind() : function (n) { for (var e = 1; e < arguments.length; e++) { var t = arguments[e]; for (var r in t) ({}).hasOwnProperty.call(t, r) && (n[r] = t[r]); } return n; }, _extends.apply(null, arguments); }
function EmptyState({
  icon = 'file-search',
  title,
  children,
  action,
  compact,
  style,
  ...rest
}) {
  return /*#__PURE__*/React.createElement("div", _extends({
    style: {
      display: 'flex',
      flexDirection: 'column',
      alignItems: 'center',
      textAlign: 'center',
      gap: 'var(--space-5)',
      padding: compact ? 'var(--space-9) var(--space-8)' : 'var(--space-12) var(--space-9)',
      border: '1px dashed var(--border-default)',
      borderRadius: 'var(--radius-md)',
      background: 'var(--surface-sunken)',
      ...style
    }
  }, rest), /*#__PURE__*/React.createElement("span", {
    style: {
      color: 'var(--text-faint)'
    }
  }, /*#__PURE__*/React.createElement(__ds_scope.Icon, {
    name: icon,
    size: compact ? 20 : 26
  })), title && /*#__PURE__*/React.createElement("div", {
    style: {
      fontFamily: 'var(--font-display)',
      fontSize: 'var(--text-h3)',
      color: 'var(--text-display)'
    }
  }, title), children && /*#__PURE__*/React.createElement("p", {
    style: {
      margin: 0,
      maxWidth: '42ch',
      fontSize: 'var(--text-body-sm)',
      color: 'var(--text-muted)',
      lineHeight: 'var(--lh-relaxed)'
    }
  }, children), action);
}
Object.assign(__ds_scope, { EmptyState });
})(); } catch (e) { __ds_ns.__errors.push({ path: "components/feedback/EmptyState.jsx", error: String((e && e.message) || e) }); }

// components/feedback/Toast.jsx
try { (() => {
function _extends() { return _extends = Object.assign ? Object.assign.bind() : function (n) { for (var e = 1; e < arguments.length; e++) { var t = arguments[e]; for (var r in t) ({}).hasOwnProperty.call(t, r) && (n[r] = t[r]); } return n; }, _extends.apply(null, arguments); }
const tones = {
  info: ['var(--info)', 'info'],
  success: ['var(--success)', 'check-check'],
  warning: ['var(--warning)', 'triangle-alert'],
  danger: ['var(--danger)', 'octagon-alert']
};
function Toast({
  tone = 'info',
  title,
  children,
  action,
  onDismiss,
  style,
  ...rest
}) {
  const [color, icon] = tones[tone] || tones.info;
  return /*#__PURE__*/React.createElement("div", _extends({
    role: "status",
    style: {
      display: 'flex',
      alignItems: 'flex-start',
      gap: 'var(--space-5)',
      width: 360,
      maxWidth: '100%',
      padding: 'var(--space-5) var(--space-6)',
      background: 'var(--surface-raised)',
      border: '1px solid var(--border-subtle)',
      borderLeft: `3px solid ${color}`,
      borderRadius: 'var(--radius-sm)',
      boxShadow: 'var(--shadow-lg)',
      ...style
    }
  }, rest), /*#__PURE__*/React.createElement("span", {
    style: {
      color,
      display: 'flex',
      marginTop: 2
    }
  }, /*#__PURE__*/React.createElement(__ds_scope.Icon, {
    name: icon,
    size: 15
  })), /*#__PURE__*/React.createElement("div", {
    style: {
      flex: 1,
      minWidth: 0
    }
  }, title && /*#__PURE__*/React.createElement("div", {
    style: {
      fontFamily: 'var(--font-display)',
      fontSize: 'var(--text-h4)',
      color: 'var(--text-display)'
    }
  }, title), children && /*#__PURE__*/React.createElement("div", {
    style: {
      fontSize: 'var(--text-caption)',
      color: 'var(--text-muted)',
      marginTop: 2,
      lineHeight: 1.5
    }
  }, children), action && /*#__PURE__*/React.createElement("div", {
    style: {
      marginTop: 'var(--space-5)'
    }
  }, action)), onDismiss && /*#__PURE__*/React.createElement(__ds_scope.IconButton, {
    icon: "x",
    label: "Dismiss",
    size: "sm",
    onClick: onDismiss
  }));
}
Object.assign(__ds_scope, { Toast });
})(); } catch (e) { __ds_ns.__errors.push({ path: "components/feedback/Toast.jsx", error: String((e && e.message) || e) }); }

// components/forms/Checkbox.jsx
try { (() => {
function _extends() { return _extends = Object.assign ? Object.assign.bind() : function (n) { for (var e = 1; e < arguments.length; e++) { var t = arguments[e]; for (var r in t) ({}).hasOwnProperty.call(t, r) && (n[r] = t[r]); } return n; }, _extends.apply(null, arguments); }
function Checkbox({
  label,
  description,
  checked,
  indeterminate,
  disabled,
  onChange,
  style,
  ...rest
}) {
  const on = checked || indeterminate;
  return /*#__PURE__*/React.createElement("label", _extends({
    style: {
      display: 'flex',
      gap: 'var(--space-5)',
      alignItems: description ? 'flex-start' : 'center',
      cursor: disabled ? 'not-allowed' : 'pointer',
      opacity: disabled ? 0.5 : 1,
      ...style
    }
  }, rest), /*#__PURE__*/React.createElement("span", {
    style: {
      width: 16,
      height: 16,
      flex: 'none',
      marginTop: description ? 2 : 0,
      borderRadius: 'var(--radius-xs)',
      display: 'grid',
      placeItems: 'center',
      background: on ? 'var(--accent)' : 'var(--surface-raised)',
      border: `1px solid ${on ? 'var(--accent)' : 'var(--border-strong)'}`,
      boxShadow: on ? 'none' : 'var(--shadow-inset)',
      color: 'var(--accent-fg)',
      transition: 'background var(--dur-fast) var(--ease-standard), border-color var(--dur-fast) var(--ease-standard)'
    }
  }, indeterminate ? /*#__PURE__*/React.createElement(__ds_scope.Icon, {
    name: "minus",
    size: 11
  }) : checked ? /*#__PURE__*/React.createElement(__ds_scope.Icon, {
    name: "check",
    size: 11
  }) : null), /*#__PURE__*/React.createElement("input", {
    type: "checkbox",
    checked: !!checked,
    disabled: disabled,
    onChange: onChange,
    style: {
      position: 'absolute',
      opacity: 0,
      width: 0,
      height: 0
    }
  }), (label || description) && /*#__PURE__*/React.createElement("span", {
    style: {
      display: 'flex',
      flexDirection: 'column',
      gap: 1
    }
  }, label && /*#__PURE__*/React.createElement("span", {
    style: {
      fontSize: 'var(--text-body-sm)',
      color: 'var(--text-primary)'
    }
  }, label), description && /*#__PURE__*/React.createElement("span", {
    style: {
      fontSize: 'var(--text-micro)',
      color: 'var(--text-muted)'
    }
  }, description)));
}
Object.assign(__ds_scope, { Checkbox });
})(); } catch (e) { __ds_ns.__errors.push({ path: "components/forms/Checkbox.jsx", error: String((e && e.message) || e) }); }

// components/forms/Field.jsx
try { (() => {
function _extends() { return _extends = Object.assign ? Object.assign.bind() : function (n) { for (var e = 1; e < arguments.length; e++) { var t = arguments[e]; for (var r in t) ({}).hasOwnProperty.call(t, r) && (n[r] = t[r]); } return n; }, _extends.apply(null, arguments); }
function Field({
  label,
  hint,
  error,
  required,
  htmlFor,
  children,
  style,
  ...rest
}) {
  return /*#__PURE__*/React.createElement("div", _extends({
    style: {
      display: 'flex',
      flexDirection: 'column',
      gap: 'var(--space-3)',
      ...style
    }
  }, rest), label && /*#__PURE__*/React.createElement("label", {
    htmlFor: htmlFor,
    style: {
      fontFamily: 'var(--font-body)',
      fontSize: 'var(--text-caption)',
      fontWeight: 'var(--weight-medium)',
      color: 'var(--text-secondary)',
      letterSpacing: 'var(--tracking-wide)'
    }
  }, label, required ? /*#__PURE__*/React.createElement("span", {
    style: {
      color: 'var(--danger)',
      marginLeft: 3
    }
  }, "*") : null), children, (error || hint) && /*#__PURE__*/React.createElement("p", {
    style: {
      margin: 0,
      fontSize: 'var(--text-micro)',
      lineHeight: 1.5,
      color: error ? 'var(--danger)' : 'var(--text-muted)',
      fontStyle: error ? 'normal' : 'italic'
    }
  }, error || hint));
}
Object.assign(__ds_scope, { Field });
})(); } catch (e) { __ds_ns.__errors.push({ path: "components/forms/Field.jsx", error: String((e && e.message) || e) }); }

// components/forms/Input.jsx
try { (() => {
function _extends() { return _extends = Object.assign ? Object.assign.bind() : function (n) { for (var e = 1; e < arguments.length; e++) { var t = arguments[e]; for (var r in t) ({}).hasOwnProperty.call(t, r) && (n[r] = t[r]); } return n; }, _extends.apply(null, arguments); }
function Input({
  iconLeft,
  suffix,
  size = 'md',
  invalid,
  mono,
  multiline,
  rows = 3,
  style,
  ...rest
}) {
  const [focus, setFocus] = React.useState(false);
  const h = size === 'sm' ? 'var(--control-h-sm)' : size === 'lg' ? 'var(--control-h-lg)' : 'var(--control-h-md)';
  const Tag = multiline ? 'textarea' : 'input';
  return /*#__PURE__*/React.createElement("div", {
    style: {
      display: 'flex',
      alignItems: multiline ? 'flex-start' : 'center',
      gap: 7,
      height: multiline ? undefined : h,
      padding: multiline ? 'var(--space-4) 10px' : '0 10px',
      background: 'var(--surface-raised)',
      borderRadius: 'var(--radius-sm)',
      border: `1px solid ${invalid ? 'var(--danger)' : focus ? 'var(--border-focus)' : 'var(--border-default)'}`,
      boxShadow: focus ? 'var(--ring-focus)' : 'var(--shadow-inset)',
      transition: 'border-color var(--dur-fast) var(--ease-standard), box-shadow var(--dur-fast) var(--ease-standard)',
      ...style
    }
  }, iconLeft ? /*#__PURE__*/React.createElement("span", {
    style: {
      color: 'var(--text-faint)',
      display: 'flex',
      marginTop: multiline ? 3 : 0
    }
  }, /*#__PURE__*/React.createElement(__ds_scope.Icon, {
    name: iconLeft,
    size: 14
  })) : null, /*#__PURE__*/React.createElement(Tag, _extends({
    rows: multiline ? rows : undefined,
    onFocus: () => setFocus(true),
    onBlur: () => setFocus(false),
    style: {
      flex: 1,
      minWidth: 0,
      border: 0,
      outline: 'none',
      background: 'transparent',
      color: 'var(--text-primary)',
      fontFamily: mono ? 'var(--font-mono)' : 'var(--font-body)',
      fontSize: size === 'sm' ? 'var(--text-caption)' : 'var(--text-body-sm)',
      lineHeight: multiline ? 1.55 : undefined,
      resize: multiline ? 'vertical' : undefined
    }
  }, rest)), suffix ? /*#__PURE__*/React.createElement("span", {
    style: {
      color: 'var(--text-faint)',
      fontSize: 'var(--text-micro)',
      fontFamily: 'var(--font-mono)',
      whiteSpace: 'nowrap'
    }
  }, suffix) : null);
}
Object.assign(__ds_scope, { Input });
})(); } catch (e) { __ds_ns.__errors.push({ path: "components/forms/Input.jsx", error: String((e && e.message) || e) }); }

// components/forms/Radio.jsx
try { (() => {
function _extends() { return _extends = Object.assign ? Object.assign.bind() : function (n) { for (var e = 1; e < arguments.length; e++) { var t = arguments[e]; for (var r in t) ({}).hasOwnProperty.call(t, r) && (n[r] = t[r]); } return n; }, _extends.apply(null, arguments); }
function Radio({
  label,
  description,
  checked,
  disabled,
  name,
  onChange,
  style,
  ...rest
}) {
  return /*#__PURE__*/React.createElement("label", _extends({
    style: {
      display: 'flex',
      gap: 'var(--space-5)',
      alignItems: description ? 'flex-start' : 'center',
      cursor: disabled ? 'not-allowed' : 'pointer',
      opacity: disabled ? 0.5 : 1,
      ...style
    }
  }, rest), /*#__PURE__*/React.createElement("span", {
    style: {
      width: 16,
      height: 16,
      flex: 'none',
      marginTop: description ? 2 : 0,
      borderRadius: '50%',
      display: 'grid',
      placeItems: 'center',
      background: 'var(--surface-raised)',
      border: `1px solid ${checked ? 'var(--accent)' : 'var(--border-strong)'}`,
      boxShadow: checked ? 'none' : 'var(--shadow-inset)',
      transition: 'border-color var(--dur-fast) var(--ease-standard)'
    }
  }, checked ? /*#__PURE__*/React.createElement("span", {
    style: {
      width: 8,
      height: 8,
      borderRadius: '50%',
      background: 'var(--accent)'
    }
  }) : null), /*#__PURE__*/React.createElement("input", {
    type: "radio",
    name: name,
    checked: !!checked,
    disabled: disabled,
    onChange: onChange,
    style: {
      position: 'absolute',
      opacity: 0,
      width: 0,
      height: 0
    }
  }), (label || description) && /*#__PURE__*/React.createElement("span", {
    style: {
      display: 'flex',
      flexDirection: 'column',
      gap: 1
    }
  }, label && /*#__PURE__*/React.createElement("span", {
    style: {
      fontSize: 'var(--text-body-sm)',
      color: 'var(--text-primary)'
    }
  }, label), description && /*#__PURE__*/React.createElement("span", {
    style: {
      fontSize: 'var(--text-micro)',
      color: 'var(--text-muted)'
    }
  }, description)));
}
Object.assign(__ds_scope, { Radio });
})(); } catch (e) { __ds_ns.__errors.push({ path: "components/forms/Radio.jsx", error: String((e && e.message) || e) }); }

// components/forms/Select.jsx
try { (() => {
function _extends() { return _extends = Object.assign ? Object.assign.bind() : function (n) { for (var e = 1; e < arguments.length; e++) { var t = arguments[e]; for (var r in t) ({}).hasOwnProperty.call(t, r) && (n[r] = t[r]); } return n; }, _extends.apply(null, arguments); }
function Select({
  options = [],
  size = 'md',
  style,
  ...rest
}) {
  const [focus, setFocus] = React.useState(false);
  const h = size === 'sm' ? 'var(--control-h-sm)' : size === 'lg' ? 'var(--control-h-lg)' : 'var(--control-h-md)';
  return /*#__PURE__*/React.createElement("div", {
    style: {
      position: 'relative',
      display: 'inline-flex',
      width: '100%',
      ...style
    }
  }, /*#__PURE__*/React.createElement("select", _extends({
    onFocus: () => setFocus(true),
    onBlur: () => setFocus(false),
    style: {
      appearance: 'none',
      width: '100%',
      height: h,
      padding: '0 30px 0 10px',
      background: 'var(--surface-raised)',
      color: 'var(--text-primary)',
      fontFamily: 'var(--font-body)',
      fontSize: size === 'sm' ? 'var(--text-caption)' : 'var(--text-body-sm)',
      border: `1px solid ${focus ? 'var(--border-focus)' : 'var(--border-default)'}`,
      borderRadius: 'var(--radius-sm)',
      boxShadow: focus ? 'var(--ring-focus)' : 'var(--shadow-inset)',
      outline: 'none',
      cursor: 'pointer'
    }
  }, rest), options.map(o => {
    const v = typeof o === 'string' ? o : o.value;
    const l = typeof o === 'string' ? o : o.label;
    return /*#__PURE__*/React.createElement("option", {
      key: v,
      value: v
    }, l);
  })), /*#__PURE__*/React.createElement("span", {
    style: {
      position: 'absolute',
      right: 9,
      top: '50%',
      transform: 'translateY(-50%)',
      color: 'var(--text-faint)',
      pointerEvents: 'none',
      display: 'flex'
    }
  }, /*#__PURE__*/React.createElement(__ds_scope.Icon, {
    name: "chevron-down",
    size: 14
  })));
}
Object.assign(__ds_scope, { Select });
})(); } catch (e) { __ds_ns.__errors.push({ path: "components/forms/Select.jsx", error: String((e && e.message) || e) }); }

// components/forms/Switch.jsx
try { (() => {
function _extends() { return _extends = Object.assign ? Object.assign.bind() : function (n) { for (var e = 1; e < arguments.length; e++) { var t = arguments[e]; for (var r in t) ({}).hasOwnProperty.call(t, r) && (n[r] = t[r]); } return n; }, _extends.apply(null, arguments); }
function Switch({
  checked,
  disabled,
  label,
  onChange,
  style,
  ...rest
}) {
  return /*#__PURE__*/React.createElement("label", _extends({
    style: {
      display: 'inline-flex',
      alignItems: 'center',
      gap: 'var(--space-5)',
      cursor: disabled ? 'not-allowed' : 'pointer',
      opacity: disabled ? 0.5 : 1,
      ...style
    }
  }, rest), /*#__PURE__*/React.createElement("span", {
    style: {
      width: 32,
      height: 18,
      borderRadius: 'var(--radius-pill)',
      flex: 'none',
      position: 'relative',
      background: checked ? 'var(--accent)' : 'var(--paper-300)',
      transition: 'background var(--dur-normal) var(--ease-standard)'
    }
  }, /*#__PURE__*/React.createElement("span", {
    style: {
      position: 'absolute',
      top: 2,
      left: checked ? 16 : 2,
      width: 14,
      height: 14,
      borderRadius: '50%',
      background: 'var(--paper-0)',
      boxShadow: 'var(--shadow-sm)',
      transition: 'left var(--dur-normal) var(--ease-standard)'
    }
  })), /*#__PURE__*/React.createElement("input", {
    type: "checkbox",
    role: "switch",
    checked: !!checked,
    disabled: disabled,
    onChange: onChange,
    style: {
      position: 'absolute',
      opacity: 0,
      width: 0,
      height: 0
    }
  }), label && /*#__PURE__*/React.createElement("span", {
    style: {
      fontSize: 'var(--text-body-sm)',
      color: 'var(--text-primary)'
    }
  }, label));
}
Object.assign(__ds_scope, { Switch });
})(); } catch (e) { __ds_ns.__errors.push({ path: "components/forms/Switch.jsx", error: String((e && e.message) || e) }); }

// components/navigation/Breadcrumbs.jsx
try { (() => {
function _extends() { return _extends = Object.assign ? Object.assign.bind() : function (n) { for (var e = 1; e < arguments.length; e++) { var t = arguments[e]; for (var r in t) ({}).hasOwnProperty.call(t, r) && (n[r] = t[r]); } return n; }, _extends.apply(null, arguments); }
function Breadcrumbs({
  items = [],
  style,
  ...rest
}) {
  return /*#__PURE__*/React.createElement("nav", _extends({
    style: {
      display: 'flex',
      alignItems: 'center',
      gap: 'var(--space-3)',
      flexWrap: 'wrap',
      ...style
    }
  }, rest), items.map((it, i) => {
    const last = i === items.length - 1;
    const label = typeof it === 'string' ? it : it.label;
    return /*#__PURE__*/React.createElement(React.Fragment, {
      key: i
    }, /*#__PURE__*/React.createElement("a", {
      href: typeof it === 'string' ? undefined : it.href || '#',
      style: {
        fontSize: 'var(--text-caption)',
        textDecoration: 'none',
        color: last ? 'var(--text-secondary)' : 'var(--text-link)',
        pointerEvents: last ? 'none' : undefined
      }
    }, label), !last && /*#__PURE__*/React.createElement("span", {
      style: {
        color: 'var(--text-faint)',
        display: 'flex'
      }
    }, /*#__PURE__*/React.createElement(__ds_scope.Icon, {
      name: "chevron-right",
      size: 12
    })));
  }));
}
Object.assign(__ds_scope, { Breadcrumbs });
})(); } catch (e) { __ds_ns.__errors.push({ path: "components/navigation/Breadcrumbs.jsx", error: String((e && e.message) || e) }); }

// components/navigation/SidebarNav.jsx
try { (() => {
function _extends() { return _extends = Object.assign ? Object.assign.bind() : function (n) { for (var e = 1; e < arguments.length; e++) { var t = arguments[e]; for (var r in t) ({}).hasOwnProperty.call(t, r) && (n[r] = t[r]); } return n; }, _extends.apply(null, arguments); }
function SidebarNav({
  groups = [],
  value,
  onChange,
  style,
  ...rest
}) {
  return /*#__PURE__*/React.createElement("nav", _extends({
    style: {
      display: 'flex',
      flexDirection: 'column',
      gap: 'var(--space-8)',
      ...style
    }
  }, rest), groups.map((g, gi) => /*#__PURE__*/React.createElement("div", {
    key: gi,
    style: {
      display: 'flex',
      flexDirection: 'column',
      gap: 'var(--space-1)'
    }
  }, g.label && /*#__PURE__*/React.createElement("div", {
    style: {
      fontSize: 'var(--text-micro)',
      letterSpacing: 'var(--tracking-caps)',
      textTransform: 'uppercase',
      color: 'var(--text-faint)',
      fontWeight: 'var(--weight-semibold)',
      padding: '0 10px var(--space-4)'
    }
  }, g.label), g.items.map(it => {
    const active = it.id === value;
    return /*#__PURE__*/React.createElement("button", {
      key: it.id,
      onClick: () => onChange && onChange(it.id),
      style: {
        display: 'flex',
        alignItems: 'center',
        gap: 'var(--space-5)',
        width: '100%',
        padding: '7px 10px',
        border: 0,
        cursor: 'pointer',
        textAlign: 'left',
        borderRadius: 'var(--radius-sm)',
        background: active ? 'var(--surface-selected)' : 'transparent',
        color: active ? 'var(--text-primary)' : 'var(--text-secondary)',
        fontFamily: 'var(--font-body)',
        fontSize: 'var(--text-body-sm)',
        fontWeight: active ? 'var(--weight-semibold)' : 'var(--weight-regular)',
        transition: 'background var(--dur-fast) var(--ease-standard)'
      }
    }, it.icon ? /*#__PURE__*/React.createElement(__ds_scope.Icon, {
      name: it.icon,
      size: 15,
      strokeWeight: active ? 'regular' : 'light'
    }) : null, /*#__PURE__*/React.createElement("span", {
      style: {
        flex: 1,
        overflow: 'hidden',
        textOverflow: 'ellipsis',
        whiteSpace: 'nowrap'
      }
    }, it.label), it.count != null && /*#__PURE__*/React.createElement("span", {
      style: {
        fontFamily: 'var(--font-mono)',
        fontSize: 'var(--text-micro)',
        color: 'var(--text-faint)'
      }
    }, it.count));
  }))));
}
Object.assign(__ds_scope, { SidebarNav });
})(); } catch (e) { __ds_ns.__errors.push({ path: "components/navigation/SidebarNav.jsx", error: String((e && e.message) || e) }); }

// components/navigation/Tabs.jsx
try { (() => {
function _extends() { return _extends = Object.assign ? Object.assign.bind() : function (n) { for (var e = 1; e < arguments.length; e++) { var t = arguments[e]; for (var r in t) ({}).hasOwnProperty.call(t, r) && (n[r] = t[r]); } return n; }, _extends.apply(null, arguments); }
function Tabs({
  tabs = [],
  value,
  onChange,
  style,
  ...rest
}) {
  return /*#__PURE__*/React.createElement("div", _extends({
    role: "tablist",
    style: {
      display: 'flex',
      gap: 'var(--space-7)',
      borderBottom: '1px solid var(--border-subtle)',
      ...style
    }
  }, rest), tabs.map(t => {
    const id = typeof t === 'string' ? t : t.id;
    const label = typeof t === 'string' ? t : t.label;
    const active = id === value;
    return /*#__PURE__*/React.createElement("button", {
      key: id,
      role: "tab",
      "aria-selected": active,
      onClick: () => onChange && onChange(id),
      style: {
        display: 'inline-flex',
        alignItems: 'center',
        gap: 6,
        padding: '0 1px var(--space-4)',
        background: 'transparent',
        border: 0,
        cursor: 'pointer',
        borderBottom: `2px solid ${active ? 'var(--accent-line)' : 'transparent'}`,
        marginBottom: -1,
        fontFamily: 'var(--font-display)',
        fontSize: 'var(--text-h4)',
        fontWeight: active ? 'var(--weight-semibold)' : 'var(--weight-regular)',
        color: active ? 'var(--text-display)' : 'var(--text-muted)',
        transition: 'color var(--dur-fast) var(--ease-standard), border-color var(--dur-fast) var(--ease-standard)'
      }
    }, typeof t !== 'string' && t.icon ? /*#__PURE__*/React.createElement(__ds_scope.Icon, {
      name: t.icon,
      size: 15
    }) : null, label, typeof t !== 'string' && t.count != null ? /*#__PURE__*/React.createElement("span", {
      style: {
        fontFamily: 'var(--font-mono)',
        fontSize: 'var(--text-micro)',
        color: 'var(--text-faint)'
      }
    }, t.count) : null);
  }));
}
Object.assign(__ds_scope, { Tabs });
})(); } catch (e) { __ds_ns.__errors.push({ path: "components/navigation/Tabs.jsx", error: String((e && e.message) || e) }); }

// components/research/EvidenceBadge.jsx
try { (() => {
function _extends() { return _extends = Object.assign ? Object.assign.bind() : function (n) { for (var e = 1; e < arguments.length; e++) { var t = arguments[e]; for (var r in t) ({}).hasOwnProperty.call(t, r) && (n[r] = t[r]); } return n; }, _extends.apply(null, arguments); }
const grades = {
  proven: {
    label: 'Proven',
    color: 'var(--evidence-proven)',
    bg: 'var(--evidence-proven-soft)',
    icon: 'shield-check'
  },
  probable: {
    label: 'Probable',
    color: 'var(--evidence-probable)',
    bg: 'var(--evidence-probable-soft)',
    icon: 'circle-check'
  },
  possible: {
    label: 'Possible',
    color: 'var(--evidence-possible)',
    bg: 'var(--evidence-possible-soft)',
    icon: 'circle-help'
  },
  disputed: {
    label: 'Disputed',
    color: 'var(--evidence-disputed)',
    bg: 'var(--evidence-disputed-soft)',
    icon: 'circle-slash'
  },
  undocumented: {
    label: 'Undocumented',
    color: 'var(--evidence-undocumented)',
    bg: 'var(--evidence-undocumented-soft)',
    icon: 'circle-dashed'
  }
};
function EvidenceBadge({
  grade = 'possible',
  sources,
  label,
  showLabel = true,
  style,
  ...rest
}) {
  const g = grades[grade] || grades.possible;
  return /*#__PURE__*/React.createElement("span", _extends({
    title: `${g.label}${sources != null ? ` · ${sources} source${sources === 1 ? '' : 's'}` : ''}`,
    style: {
      display: 'inline-flex',
      alignItems: 'center',
      gap: 5,
      padding: showLabel ? '2px 8px 2px 6px' : 3,
      background: g.bg,
      color: g.color,
      border: '1px solid transparent',
      borderRadius: 'var(--radius-xs)',
      fontFamily: 'var(--font-body)',
      fontSize: 'var(--text-micro)',
      fontWeight: 'var(--weight-semibold)',
      letterSpacing: 'var(--tracking-caps)',
      textTransform: 'uppercase',
      whiteSpace: 'nowrap',
      ...style
    }
  }, rest), /*#__PURE__*/React.createElement(__ds_scope.Icon, {
    name: g.icon,
    size: 12
  }), showLabel ? label || g.label : null, sources != null && showLabel ? /*#__PURE__*/React.createElement("span", {
    style: {
      fontFamily: 'var(--font-mono)',
      letterSpacing: 0,
      opacity: 0.8
    }
  }, sources) : null);
}
Object.assign(__ds_scope, { EvidenceBadge });
})(); } catch (e) { __ds_ns.__errors.push({ path: "components/research/EvidenceBadge.jsx", error: String((e && e.message) || e) }); }

// components/research/FactRow.jsx
try { (() => {
function _extends() { return _extends = Object.assign ? Object.assign.bind() : function (n) { for (var e = 1; e < arguments.length; e++) { var t = arguments[e]; for (var r in t) ({}).hasOwnProperty.call(t, r) && (n[r] = t[r]); } return n; }, _extends.apply(null, arguments); }
function FactRow({
  type = 'birth',
  label,
  value,
  place,
  date,
  sources,
  grade = 'probable',
  conflict,
  actions,
  style,
  ...rest
}) {
  const [hover, setHover] = React.useState(false);
  const color = `var(--rec-${type})`;
  const icons = {
    birth: 'baby',
    marriage: 'heart-handshake',
    death: 'cross',
    census: 'table-2',
    migration: 'ship',
    military: 'shield',
    probate: 'gavel',
    dna: 'dna'
  };
  return /*#__PURE__*/React.createElement("div", _extends({
    onMouseEnter: () => setHover(true),
    onMouseLeave: () => setHover(false),
    style: {
      display: 'grid',
      gridTemplateColumns: '22px 128px 1fr auto',
      alignItems: 'baseline',
      gap: 'var(--space-6)',
      padding: 'var(--space-5) var(--space-6)',
      background: hover ? 'var(--surface-hover)' : 'transparent',
      borderBottom: '1px solid var(--border-subtle)',
      transition: 'background var(--dur-fast) var(--ease-standard)',
      ...style
    }
  }, rest), /*#__PURE__*/React.createElement("span", {
    style: {
      color,
      display: 'flex',
      alignSelf: 'center'
    }
  }, /*#__PURE__*/React.createElement(__ds_scope.Icon, {
    name: icons[type] || 'file-text',
    size: 15
  })), /*#__PURE__*/React.createElement("span", {
    style: {
      fontSize: 'var(--text-caption)',
      letterSpacing: 'var(--tracking-wide)',
      color: 'var(--text-muted)',
      textTransform: 'capitalize'
    }
  }, label || type), /*#__PURE__*/React.createElement("span", {
    style: {
      display: 'flex',
      flexDirection: 'column',
      gap: 2,
      minWidth: 0
    }
  }, /*#__PURE__*/React.createElement("span", {
    style: {
      fontSize: 'var(--text-body-sm)',
      color: 'var(--text-primary)'
    }
  }, date && /*#__PURE__*/React.createElement("span", {
    style: {
      fontFamily: 'var(--font-mono)',
      fontSize: 'var(--text-caption)',
      marginRight: 8,
      color: 'var(--text-secondary)'
    }
  }, date), value), place && /*#__PURE__*/React.createElement("span", {
    style: {
      fontSize: 'var(--text-micro)',
      color: 'var(--text-muted)',
      fontStyle: 'italic'
    }
  }, place), conflict && /*#__PURE__*/React.createElement("span", {
    style: {
      display: 'inline-flex',
      alignItems: 'center',
      gap: 4,
      fontSize: 'var(--text-micro)',
      color: 'var(--danger)'
    }
  }, /*#__PURE__*/React.createElement(__ds_scope.Icon, {
    name: "git-compare-arrows",
    size: 12
  }), conflict)), /*#__PURE__*/React.createElement("span", {
    style: {
      display: 'flex',
      alignItems: 'center',
      gap: 'var(--space-4)',
      alignSelf: 'center'
    }
  }, /*#__PURE__*/React.createElement(__ds_scope.EvidenceBadge, {
    grade: grade,
    sources: sources
  }), actions));
}
Object.assign(__ds_scope, { FactRow });
})(); } catch (e) { __ds_ns.__errors.push({ path: "components/research/FactRow.jsx", error: String((e && e.message) || e) }); }

// components/research/PersonChip.jsx
try { (() => {
function _extends() { return _extends = Object.assign ? Object.assign.bind() : function (n) { for (var e = 1; e < arguments.length; e++) { var t = arguments[e]; for (var r in t) ({}).hasOwnProperty.call(t, r) && (n[r] = t[r]); } return n; }, _extends.apply(null, arguments); }
function PersonChip({
  name,
  years,
  line = 'paternal',
  sex,
  avatar,
  grade,
  size = 'md',
  selected,
  onClick,
  style,
  ...rest
}) {
  const [hover, setHover] = React.useState(false);
  const lineColor = line === 'maternal' ? 'var(--line-maternal)' : line === 'inferred' ? 'var(--line-inferred)' : 'var(--line-paternal)';
  const d = size === 'sm' ? 26 : 34;
  return /*#__PURE__*/React.createElement("div", _extends({
    onClick: onClick,
    onMouseEnter: () => setHover(true),
    onMouseLeave: () => setHover(false),
    style: {
      display: 'inline-flex',
      alignItems: 'center',
      gap: 'var(--space-5)',
      padding: size === 'sm' ? '4px 10px 4px 5px' : '6px 12px 6px 6px',
      background: selected ? 'var(--surface-selected)' : 'var(--surface-card)',
      border: `1px solid ${selected ? 'var(--accent-line)' : 'var(--border-subtle)'}`,
      borderLeft: `3px solid ${lineColor}`,
      borderRadius: 'var(--radius-sm)',
      boxShadow: hover && onClick ? 'var(--shadow-md)' : 'var(--shadow-sm)',
      cursor: onClick ? 'pointer' : 'default',
      minWidth: 0,
      transition: 'box-shadow var(--dur-fast) var(--ease-standard), background var(--dur-fast) var(--ease-standard)',
      ...style
    }
  }, rest), /*#__PURE__*/React.createElement("span", {
    style: {
      width: d,
      height: d,
      flex: 'none',
      borderRadius: 'var(--radius-xs)',
      display: 'grid',
      placeItems: 'center',
      background: avatar ? `center/cover no-repeat url(${avatar})` : 'var(--surface-inset)',
      color: 'var(--text-faint)',
      fontFamily: 'var(--font-display)',
      fontSize: 13
    }
  }, !avatar && /*#__PURE__*/React.createElement(__ds_scope.Icon, {
    name: sex === 'F' ? 'user-round' : 'user',
    size: 14
  })), /*#__PURE__*/React.createElement("span", {
    style: {
      display: 'flex',
      flexDirection: 'column',
      minWidth: 0
    }
  }, /*#__PURE__*/React.createElement("span", {
    style: {
      fontFamily: 'var(--font-display)',
      fontSize: size === 'sm' ? 'var(--text-body-sm)' : 'var(--text-h4)',
      fontWeight: 'var(--weight-medium)',
      color: 'var(--text-display)',
      whiteSpace: 'nowrap',
      overflow: 'hidden',
      textOverflow: 'ellipsis'
    }
  }, name), years && /*#__PURE__*/React.createElement("span", {
    style: {
      fontFamily: 'var(--font-mono)',
      fontSize: 'var(--text-micro)',
      color: 'var(--text-muted)'
    }
  }, years)), grade === 'disputed' && /*#__PURE__*/React.createElement(__ds_scope.Icon, {
    name: "circle-slash",
    size: 13,
    style: {
      color: 'var(--evidence-disputed)'
    }
  }), grade === 'possible' && /*#__PURE__*/React.createElement(__ds_scope.Icon, {
    name: "circle-help",
    size: 13,
    style: {
      color: 'var(--evidence-possible)'
    }
  }));
}
Object.assign(__ds_scope, { PersonChip });
})(); } catch (e) { __ds_ns.__errors.push({ path: "components/research/PersonChip.jsx", error: String((e && e.message) || e) }); }

// components/research/SourceCitation.jsx
try { (() => {
function _extends() { return _extends = Object.assign ? Object.assign.bind() : function (n) { for (var e = 1; e < arguments.length; e++) { var t = arguments[e]; for (var r in t) ({}).hasOwnProperty.call(t, r) && (n[r] = t[r]); } return n; }, _extends.apply(null, arguments); }
function SourceCitation({
  citation,
  repository,
  accessed,
  grade,
  image,
  quality,
  actions,
  style,
  ...rest
}) {
  return /*#__PURE__*/React.createElement("article", _extends({
    style: {
      display: 'flex',
      gap: 'var(--space-6)',
      padding: 'var(--space-6)',
      background: 'var(--surface-card)',
      border: '1px solid var(--border-subtle)',
      borderRadius: 'var(--radius-sm)',
      ...style
    }
  }, rest), /*#__PURE__*/React.createElement("div", {
    style: {
      width: 44,
      height: 56,
      flex: 'none',
      borderRadius: 'var(--radius-xs)',
      background: image ? `center/cover no-repeat url(${image})` : 'var(--surface-inset)',
      border: '1px solid var(--border-default)',
      display: 'grid',
      placeItems: 'center',
      color: 'var(--text-faint)'
    }
  }, !image && /*#__PURE__*/React.createElement(__ds_scope.Icon, {
    name: "scroll-text",
    size: 16
  })), /*#__PURE__*/React.createElement("div", {
    style: {
      flex: 1,
      minWidth: 0
    }
  }, /*#__PURE__*/React.createElement("p", {
    style: {
      margin: 0,
      fontFamily: 'var(--font-body)',
      fontSize: 'var(--text-body-sm)',
      lineHeight: 'var(--lh-relaxed)',
      color: 'var(--text-primary)'
    }
  }, citation), /*#__PURE__*/React.createElement("div", {
    style: {
      display: 'flex',
      alignItems: 'center',
      flexWrap: 'wrap',
      gap: 'var(--space-5)',
      marginTop: 'var(--space-4)'
    }
  }, grade && /*#__PURE__*/React.createElement(__ds_scope.EvidenceBadge, {
    grade: grade
  }), repository && /*#__PURE__*/React.createElement("span", {
    style: {
      display: 'inline-flex',
      alignItems: 'center',
      gap: 4,
      fontSize: 'var(--text-micro)',
      color: 'var(--text-muted)'
    }
  }, /*#__PURE__*/React.createElement(__ds_scope.Icon, {
    name: "library",
    size: 12
  }), repository), quality && /*#__PURE__*/React.createElement("span", {
    style: {
      fontSize: 'var(--text-micro)',
      color: 'var(--text-muted)',
      fontStyle: 'italic'
    }
  }, quality), accessed && /*#__PURE__*/React.createElement("span", {
    style: {
      fontFamily: 'var(--font-mono)',
      fontSize: 'var(--text-micro)',
      color: 'var(--text-faint)'
    }
  }, accessed))), actions && /*#__PURE__*/React.createElement("div", {
    style: {
      display: 'flex',
      gap: 2,
      alignItems: 'flex-start'
    }
  }, actions));
}
Object.assign(__ds_scope, { SourceCitation });
})(); } catch (e) { __ds_ns.__errors.push({ path: "components/research/SourceCitation.jsx", error: String((e && e.message) || e) }); }

// ui_kits/app/App.jsx
try { (() => {
const {
  Button,
  IconButton,
  Badge,
  Icon,
  Dialog,
  Toast,
  Checkbox,
  Field,
  Input,
  Select,
  Tabs,
  EvidenceBadge,
  Radio
} = window.ProvenenciaDesignSystem_0f6c1f;
function AttachDialog({
  onClose,
  onDone
}) {
  const [facts, setFacts] = React.useState(['birth', 'census']);
  const t = k => setFacts(f => f.includes(k) ? f.filter(x => x !== k) : [...f, k]);
  return /*#__PURE__*/React.createElement(Dialog, {
    width: 480,
    title: "Attach source",
    subtitle: "1881 census of England \u2014 Great Yarmouth, ED 12, sched. 41",
    onClose: onClose,
    footer: /*#__PURE__*/React.createElement(React.Fragment, null, /*#__PURE__*/React.createElement(Button, {
      variant: "ghost",
      onClick: onClose
    }, "Cancel"), /*#__PURE__*/React.createElement(Button, {
      variant: "primary",
      iconLeft: "paperclip",
      onClick: onDone
    }, "Attach to ", facts.length, " facts"))
  }, /*#__PURE__*/React.createElement("div", {
    style: {
      display: 'flex',
      flexDirection: 'column',
      gap: 'var(--space-7)'
    }
  }, /*#__PURE__*/React.createElement("div", null, /*#__PURE__*/React.createElement("div", {
    className: "eyebrow"
  }, "Which facts does this record support?"), /*#__PURE__*/React.createElement("div", {
    style: {
      display: 'flex',
      flexDirection: 'column',
      gap: 'var(--space-5)',
      marginTop: 'var(--space-5)'
    }
  }, [['birth', 'Birth · abt 1847, Norfolk', 'Age 34 stated'], ['census', 'Residence · 12 Row 62, Great Yarmouth', 'Household head'], ['occupation', 'Occupation · mariner', ''], ['marriage', 'Marriage · Ellen Marsh', 'Wife in household']].map(([k, l, d]) => /*#__PURE__*/React.createElement(Checkbox, {
    key: k,
    checked: facts.includes(k),
    onChange: () => t(k),
    label: l,
    description: d || undefined
  })))), /*#__PURE__*/React.createElement(Field, {
    label: "Evidence type"
  }, /*#__PURE__*/React.createElement("div", {
    style: {
      display: 'flex',
      gap: 'var(--space-8)'
    }
  }, /*#__PURE__*/React.createElement(Radio, {
    name: "ev",
    checked: true,
    label: "Direct"
  }), /*#__PURE__*/React.createElement(Radio, {
    name: "ev",
    label: "Indirect"
  }), /*#__PURE__*/React.createElement(Radio, {
    name: "ev",
    label: "Negative"
  }))), /*#__PURE__*/React.createElement(Field, {
    label: "Research note",
    hint: "Recorded in the log against today's date"
  }, /*#__PURE__*/React.createElement(Input, {
    multiline: true,
    rows: 2,
    defaultValue: "Household composition matches the 1871 marriage and known children."
  }))));
}
function App() {
  const [theme, setTheme] = React.useState('light');
  const [view, setView] = React.useState('dashboard');
  const [query, setQuery] = React.useState('');
  const [dialog, setDialog] = React.useState(false);
  const [toast, setToast] = React.useState(null);
  React.useEffect(() => {
    document.documentElement.setAttribute('data-theme', theme);
  }, [theme]);
  const openPerson = () => setView('person');
  const done = () => {
    setDialog(false);
    setToast('Source attached to 2 facts on Thomas Alderwick.');
    setTimeout(() => setToast(null), 4200);
  };
  const heads = {
    dashboard: {
      title: 'Overview',
      crumbs: [{
        label: 'Alderwick line',
        href: '#'
      }, {
        label: 'Overview'
      }],
      meta: /*#__PURE__*/React.createElement(React.Fragment, null, /*#__PURE__*/React.createElement("span", {
        style: {
          fontSize: 'var(--text-caption)',
          color: 'var(--text-muted)'
        }
      }, "412 people \xB7 1,284 sources \xB7 94 open conflicts")),
      actions: /*#__PURE__*/React.createElement(React.Fragment, null, /*#__PURE__*/React.createElement(Button, {
        variant: "secondary",
        iconLeft: "upload"
      }, "Import GEDCOM"), /*#__PURE__*/React.createElement(Button, {
        variant: "primary",
        iconLeft: "user-plus"
      }, "Add person"))
    },
    person: {
      title: 'Thomas Alderwick',
      crumbs: [{
        label: 'Alderwick line',
        href: '#'
      }, {
        label: 'People',
        href: '#'
      }, {
        label: 'Thomas Alderwick'
      }],
      meta: /*#__PURE__*/React.createElement(React.Fragment, null, /*#__PURE__*/React.createElement("span", {
        style: {
          fontFamily: 'var(--font-mono)',
          fontSize: 'var(--text-caption)',
          color: 'var(--text-secondary)'
        }
      }, "1847\u20131912"), /*#__PURE__*/React.createElement(EvidenceBadge, {
        grade: "probable",
        sources: 31
      }), /*#__PURE__*/React.createElement(Badge, {
        tone: "danger",
        icon: "git-compare-arrows"
      }, "2 conflicts")),
      actions: /*#__PURE__*/React.createElement(React.Fragment, null, /*#__PURE__*/React.createElement(Button, {
        variant: "secondary",
        iconLeft: "git-branch",
        onClick: () => setView('tree')
      }, "Pedigree"), /*#__PURE__*/React.createElement(Button, {
        variant: "primary",
        iconLeft: "paperclip",
        onClick: () => setDialog(true)
      }, "Attach source"))
    },
    tree: {
      title: 'Pedigree',
      crumbs: [{
        label: 'Alderwick line',
        href: '#'
      }, {
        label: 'Pedigree'
      }],
      meta: /*#__PURE__*/React.createElement("span", {
        style: {
          fontSize: 'var(--text-caption)',
          color: 'var(--text-muted)'
        }
      }, "Thomas Alderwick \xB7 4 generations \xB7 3 gaps"),
      actions: /*#__PURE__*/React.createElement(Button, {
        variant: "secondary",
        iconLeft: "user",
        onClick: openPerson
      }, "Back to person")
    },
    sources: {
      title: 'Sources',
      crumbs: [{
        label: 'Alderwick line',
        href: '#'
      }, {
        label: 'Sources'
      }],
      meta: /*#__PURE__*/React.createElement("span", {
        style: {
          fontSize: 'var(--text-caption)',
          color: 'var(--text-muted)'
        }
      }, "1,284 sources \xB7 61 repositories"),
      actions: /*#__PURE__*/React.createElement(Button, {
        variant: "primary",
        iconLeft: "plus",
        onClick: () => setDialog(true)
      }, "New source")
    }
  };
  const active = ['dashboard', 'person', 'tree', 'sources'].includes(view) ? view : 'dashboard';
  const h = heads[active];
  return /*#__PURE__*/React.createElement("div", {
    style: {
      position: 'relative',
      height: '100vh',
      display: 'flex',
      flexDirection: 'column',
      background: 'var(--surface-page)',
      overflow: 'hidden'
    }
  }, /*#__PURE__*/React.createElement(TopBar, {
    theme: theme,
    onTheme: () => setTheme(t => t === 'light' ? 'dark' : 'light'),
    query: query,
    setQuery: setQuery,
    onSearch: () => setView('sources')
  }), /*#__PURE__*/React.createElement("div", {
    style: {
      flex: 1,
      display: 'flex',
      minHeight: 0
    }
  }, /*#__PURE__*/React.createElement(Sidebar, {
    view: view === 'person' ? 'people' : view,
    onView: id => setView(['dashboard', 'tree', 'sources'].includes(id) ? id : id === 'people' ? 'person' : 'sources')
  }), /*#__PURE__*/React.createElement("main", {
    style: {
      flex: 1,
      minWidth: 0,
      display: 'flex',
      flexDirection: 'column',
      overflow: active === 'dashboard' ? 'auto' : 'hidden'
    }
  }, /*#__PURE__*/React.createElement(PageHeader, h), active === 'dashboard' && /*#__PURE__*/React.createElement(Dashboard, {
    onOpenPerson: openPerson,
    onAttach: () => setDialog(true)
  }), active === 'person' && /*#__PURE__*/React.createElement(PersonView, {
    onAttach: () => setDialog(true)
  }), active === 'tree' && /*#__PURE__*/React.createElement(TreeView, {
    onOpenPerson: openPerson
  }), active === 'sources' && /*#__PURE__*/React.createElement(SourcesLibrary, {
    onAttach: () => setDialog(true)
  }))), dialog && /*#__PURE__*/React.createElement(AttachDialog, {
    onClose: () => setDialog(false),
    onDone: done
  }), toast && /*#__PURE__*/React.createElement("div", {
    style: {
      position: 'absolute',
      left: 'var(--gutter-page)',
      bottom: 'var(--space-8)',
      zIndex: 70
    }
  }, /*#__PURE__*/React.createElement(Toast, {
    tone: "success",
    title: "Source attached",
    onDismiss: () => setToast(null)
  }, toast)));
}
const appRoot = document.getElementById('root');
if (appRoot) ReactDOM.createRoot(appRoot).render(/*#__PURE__*/React.createElement(App, null));
})(); } catch (e) { __ds_ns.__errors.push({ path: "ui_kits/app/App.jsx", error: String((e && e.message) || e) }); }

// ui_kits/app/Dashboard.jsx
try { (() => {
const {
  Card,
  Button,
  IconButton,
  Badge,
  Icon,
  EvidenceBadge,
  PersonChip,
  SourceCitation,
  EmptyState,
  Tag
} = window.ProvenenciaDesignSystem_0f6c1f;
function StatBlock({
  label,
  value,
  note,
  tone
}) {
  return /*#__PURE__*/React.createElement("div", {
    style: {
      flex: 1,
      padding: 'var(--space-6) var(--space-7)',
      background: 'var(--surface-card)',
      border: '1px solid var(--border-subtle)',
      borderRadius: 'var(--radius-md)',
      boxShadow: 'var(--shadow-sm)'
    }
  }, /*#__PURE__*/React.createElement("div", {
    className: "eyebrow"
  }, label), /*#__PURE__*/React.createElement("div", {
    style: {
      fontFamily: 'var(--font-display)',
      fontSize: 'var(--text-display-3)',
      color: tone || 'var(--text-display)',
      lineHeight: 1.1,
      marginTop: 4
    }
  }, value), /*#__PURE__*/React.createElement("div", {
    style: {
      fontSize: 'var(--text-micro)',
      color: 'var(--text-muted)',
      marginTop: 2
    }
  }, note));
}
function Dashboard({
  onOpenPerson,
  onAttach
}) {
  return /*#__PURE__*/React.createElement("div", {
    style: {
      padding: 'var(--space-8) var(--gutter-page) var(--space-12)',
      display: 'flex',
      flexDirection: 'column',
      gap: 'var(--space-8)',
      maxWidth: 'var(--width-content-max)'
    }
  }, /*#__PURE__*/React.createElement("div", {
    style: {
      display: 'flex',
      gap: 'var(--space-6)'
    }
  }, /*#__PURE__*/React.createElement(StatBlock, {
    label: "People",
    value: "412",
    note: "+14 this week"
  }), /*#__PURE__*/React.createElement(StatBlock, {
    label: "Sources cited",
    value: "1,284",
    note: "94% of facts cited"
  }), /*#__PURE__*/React.createElement(StatBlock, {
    label: "Open conflicts",
    value: "94",
    note: "18 flagged this week",
    tone: "var(--danger)"
  }), /*#__PURE__*/React.createElement(StatBlock, {
    label: "Proven facts",
    value: "52%",
    note: "of 1,904 assertions",
    tone: "var(--success)"
  })), /*#__PURE__*/React.createElement("div", {
    style: {
      display: 'grid',
      gridTemplateColumns: '1.35fr 1fr',
      gap: 'var(--space-8)',
      alignItems: 'start'
    }
  }, /*#__PURE__*/React.createElement(Card, {
    title: "Research log",
    subtitle: "Week 12 \xB7 Great Yarmouth mariners",
    actions: /*#__PURE__*/React.createElement(Button, {
      size: "sm",
      variant: "ghost",
      iconLeft: "plus"
    }, "New entry")
  }, /*#__PURE__*/React.createElement("div", {
    style: {
      display: 'flex',
      flexDirection: 'column',
      gap: 'var(--space-7)'
    }
  }, [{
    d: '14 Mar',
    t: 'Searched RG11 for Alderwick households in Great Yarmouth',
    r: 'Found sched. 41 — matches known children.',
    g: 'proven',
    tags: [['1881 Census', 'var(--rec-census)']]
  }, {
    d: '12 Mar',
    t: 'Ordered death certificate for Thomas Alderwick',
    r: 'GRO reference confirmed; awaiting scan.',
    g: 'possible',
    tags: [['Civil registration', 'var(--rec-probate)']]
  }, {
    d: '09 Mar',
    t: 'Compared ages across 1881 and 1891 censuses',
    r: 'Two-year discrepancy — recorded as a conflict.',
    g: 'disputed',
    tags: [['Conflict', 'var(--rec-dna)'], ['Age', 'var(--rec-birth)']]
  }].map((e, i) => /*#__PURE__*/React.createElement("div", {
    key: i,
    style: {
      display: 'grid',
      gridTemplateColumns: '52px 1fr',
      gap: 'var(--space-6)'
    }
  }, /*#__PURE__*/React.createElement("div", {
    style: {
      fontFamily: 'var(--font-mono)',
      fontSize: 'var(--text-micro)',
      color: 'var(--text-faint)',
      paddingTop: 3
    }
  }, e.d), /*#__PURE__*/React.createElement("div", {
    style: {
      borderLeft: '1px solid var(--border-subtle)',
      paddingLeft: 'var(--space-6)'
    }
  }, /*#__PURE__*/React.createElement("div", {
    style: {
      fontSize: 'var(--text-body-sm)',
      color: 'var(--text-primary)'
    }
  }, e.t), /*#__PURE__*/React.createElement("div", {
    style: {
      fontSize: 'var(--text-caption)',
      color: 'var(--text-muted)',
      fontStyle: 'italic',
      marginTop: 3
    }
  }, e.r), /*#__PURE__*/React.createElement("div", {
    style: {
      display: 'flex',
      gap: 'var(--space-4)',
      marginTop: 'var(--space-5)',
      alignItems: 'center'
    }
  }, /*#__PURE__*/React.createElement(EvidenceBadge, {
    grade: e.g
  }), e.tags.map(([l, c]) => /*#__PURE__*/React.createElement(Tag, {
    key: l,
    color: c
  }, l)))))))), /*#__PURE__*/React.createElement("div", {
    style: {
      display: 'flex',
      flexDirection: 'column',
      gap: 'var(--space-8)'
    }
  }, /*#__PURE__*/React.createElement(Card, {
    title: "Needs review",
    subtitle: "Highest-impact conflicts first",
    padding: "var(--space-6)"
  }, /*#__PURE__*/React.createElement("div", {
    style: {
      display: 'flex',
      flexDirection: 'column',
      gap: 'var(--space-5)'
    }
  }, /*#__PURE__*/React.createElement(PersonChip, {
    name: "Thomas Alderwick",
    years: "1847\u20131912",
    line: "paternal",
    sex: "M",
    grade: "disputed",
    onClick: onOpenPerson
  }), /*#__PURE__*/React.createElement(PersonChip, {
    name: "Ellen Marsh",
    years: "b. 1851",
    line: "maternal",
    sex: "F",
    grade: "possible",
    onClick: onOpenPerson
  }), /*#__PURE__*/React.createElement(PersonChip, {
    name: "Mary Ann Alderwick",
    years: "1873\u20131873",
    line: "paternal",
    sex: "F",
    grade: "possible",
    onClick: onOpenPerson
  }))), /*#__PURE__*/React.createElement(Card, {
    title: "Newly indexed",
    subtitle: "Matching your open questions",
    padding: "var(--space-6)"
  }, /*#__PURE__*/React.createElement(SourceCitation, {
    style: {
      border: 0,
      padding: 0,
      background: 'transparent'
    },
    citation: "Norfolk, England, Church of England marriages and banns, 1754\u20131936, Great Yarmouth St Nicholas, 1871, entry 214.",
    repository: "Norfolk Record Office",
    grade: "probable",
    quality: "Derivative \xB7 index entry",
    actions: /*#__PURE__*/React.createElement(IconButton, {
      icon: "paperclip",
      label: "Attach",
      size: "sm",
      onClick: onAttach
    })
  })))));
}
Object.assign(window, {
  Dashboard
});
})(); } catch (e) { __ds_ns.__errors.push({ path: "ui_kits/app/Dashboard.jsx", error: String((e && e.message) || e) }); }

// ui_kits/app/PersonView.jsx
try { (() => {
const {
  Card,
  Button,
  IconButton,
  Badge,
  Icon,
  Tabs,
  FactRow,
  SourceCitation,
  PersonChip,
  EvidenceBadge,
  EmptyState,
  Callout,
  Tag
} = window.ProvenenciaDesignSystem_0f6c1f;
function PersonView({
  onAttach
}) {
  const [tab, setTab] = React.useState('facts');
  return /*#__PURE__*/React.createElement("div", {
    style: {
      display: 'flex',
      gap: 0,
      alignItems: 'stretch',
      flex: 1,
      minHeight: 0
    }
  }, /*#__PURE__*/React.createElement("div", {
    style: {
      flex: 1,
      minWidth: 0,
      overflow: 'auto',
      padding: '0 var(--gutter-page) var(--space-12)'
    }
  }, /*#__PURE__*/React.createElement("div", {
    style: {
      marginTop: 'var(--space-7)'
    }
  }, /*#__PURE__*/React.createElement(Tabs, {
    value: tab,
    onChange: setTab,
    tabs: [{
      id: 'facts',
      label: 'Facts',
      icon: 'list',
      count: 12
    }, {
      id: 'sources',
      label: 'Sources',
      icon: 'scroll-text',
      count: 31
    }, {
      id: 'family',
      label: 'Family',
      icon: 'users',
      count: 8
    }]
  })), tab === 'facts' && /*#__PURE__*/React.createElement(React.Fragment, null, /*#__PURE__*/React.createElement(Callout, {
    style: {
      marginTop: 'var(--space-7)'
    },
    tone: "danger",
    title: "Two sources give different death dates",
    detail: "GRO death index, Q2 1912 · Norwich, age 65\n1891 census, age 42 · implies b. 1849",
    actions: /*#__PURE__*/React.createElement(React.Fragment, null, /*#__PURE__*/React.createElement(Button, {
      size: "sm",
      variant: "secondary",
      iconLeft: "notebook-pen"
    }, "Record a conclusion"), /*#__PURE__*/React.createElement(Button, {
      size: "sm",
      variant: "link"
    }, "View both sources"))
  }, "This identity stays at ", /*#__PURE__*/React.createElement("em", null, "probable"), " while the conflict is open."), /*#__PURE__*/React.createElement("div", {
    style: {
      marginTop: 'var(--space-6)',
      background: 'var(--surface-card)',
      border: '1px solid var(--border-subtle)',
      borderRadius: 'var(--radius-md)',
      overflow: 'hidden',
      boxShadow: 'var(--shadow-sm)'
    }
  }, /*#__PURE__*/React.createElement(FactRow, {
    type: "birth",
    date: "12 Mar 1847",
    value: "Great Yarmouth, Norfolk",
    place: "St Nicholas parish register, bapt. 4 Apr 1847",
    sources: 3,
    grade: "proven",
    actions: /*#__PURE__*/React.createElement(IconButton, {
      icon: "pencil",
      label: "Edit fact",
      size: "sm"
    })
  }), /*#__PURE__*/React.createElement(FactRow, {
    type: "census",
    date: "3 Apr 1881",
    value: "12 Row 62, Great Yarmouth",
    place: "RG11/1976, ED 12, sched. 41 \u2014 mariner",
    sources: 2,
    grade: "proven",
    actions: /*#__PURE__*/React.createElement(IconButton, {
      icon: "pencil",
      label: "Edit fact",
      size: "sm"
    })
  }), /*#__PURE__*/React.createElement(FactRow, {
    type: "marriage",
    date: "4 Jun 1871",
    value: "Ellen Marsh",
    place: "Great Yarmouth St Nicholas",
    sources: 2,
    grade: "probable",
    actions: /*#__PURE__*/React.createElement(IconButton, {
      icon: "pencil",
      label: "Edit fact",
      size: "sm"
    })
  }), /*#__PURE__*/React.createElement(FactRow, {
    type: "migration",
    date: "1883",
    value: "Voyage to Grimsby, then returned",
    place: "Crew list, SS Iceni",
    sources: 1,
    grade: "possible",
    actions: /*#__PURE__*/React.createElement(IconButton, {
      icon: "pencil",
      label: "Edit fact",
      size: "sm"
    })
  }), /*#__PURE__*/React.createElement(FactRow, {
    type: "military",
    date: "1866\u20131869",
    value: "Royal Naval Reserve, Yarmouth division",
    sources: 1,
    grade: "possible",
    actions: /*#__PURE__*/React.createElement(IconButton, {
      icon: "pencil",
      label: "Edit fact",
      size: "sm"
    })
  }), /*#__PURE__*/React.createElement(FactRow, {
    type: "death",
    date: "abt 1912",
    value: "Norwich",
    place: "GRO index, Norwich district, Q2 1912",
    sources: 1,
    grade: "disputed",
    conflict: "Conflicts with 1891 census age (implies b. 1849)",
    actions: /*#__PURE__*/React.createElement(IconButton, {
      icon: "pencil",
      label: "Edit fact",
      size: "sm"
    })
  }))), tab === 'sources' && /*#__PURE__*/React.createElement("div", {
    style: {
      marginTop: 'var(--space-7)',
      display: 'flex',
      flexDirection: 'column',
      gap: 'var(--space-5)'
    }
  }, /*#__PURE__*/React.createElement(SourceCitation, {
    citation: "1881 census of England, Norfolk, Great Yarmouth, ED 12, sched. 41, Thomas Alderwick household; digital image, The National Archives (RG11/1976), fol. 23.",
    repository: "The National Archives",
    accessed: "accessed 14 Mar 2026",
    grade: "proven",
    quality: "Original \xB7 digital image",
    actions: /*#__PURE__*/React.createElement(React.Fragment, null, /*#__PURE__*/React.createElement(IconButton, {
      icon: "quote",
      label: "Copy citation",
      size: "sm"
    }), /*#__PURE__*/React.createElement(IconButton, {
      icon: "external-link",
      label: "Open image",
      size: "sm"
    }))
  }), /*#__PURE__*/React.createElement(SourceCitation, {
    citation: "Norfolk, England, Church of England marriages and banns, 1754\u20131936, Great Yarmouth St Nicholas, 1871, entry 214.",
    repository: "Norfolk Record Office (PD 28/109)",
    accessed: "accessed 12 Mar 2026",
    grade: "probable",
    quality: "Derivative \xB7 index entry",
    actions: /*#__PURE__*/React.createElement(IconButton, {
      icon: "quote",
      label: "Copy citation",
      size: "sm"
    })
  }), /*#__PURE__*/React.createElement(SourceCitation, {
    citation: "General Register Office, death index, Norwich district, Apr\u2013Jun 1912, vol. 4b, p. 112, Thomas Alderwick, age 65.",
    repository: "GRO",
    accessed: "accessed 09 Mar 2026",
    grade: "disputed",
    quality: "Derivative \xB7 index only",
    actions: /*#__PURE__*/React.createElement(IconButton, {
      icon: "quote",
      label: "Copy citation",
      size: "sm"
    })
  }), /*#__PURE__*/React.createElement(EmptyState, {
    compact: true,
    icon: "file-plus",
    title: "Certificate ordered, not yet received",
    action: /*#__PURE__*/React.createElement(Button, {
      size: "sm",
      variant: "secondary",
      iconLeft: "plus",
      onClick: onAttach
    }, "Attach a source")
  }, "Facts stay at ", /*#__PURE__*/React.createElement("em", null, "possible"), " until the original is in hand.")), tab === 'family' && /*#__PURE__*/React.createElement("div", {
    style: {
      marginTop: 'var(--space-7)',
      display: 'grid',
      gridTemplateColumns: '1fr 1fr',
      gap: 'var(--space-8)'
    }
  }, /*#__PURE__*/React.createElement(Card, {
    title: "Parents",
    padding: "var(--space-6)"
  }, /*#__PURE__*/React.createElement("div", {
    style: {
      display: 'flex',
      flexDirection: 'column',
      gap: 'var(--space-5)'
    }
  }, /*#__PURE__*/React.createElement(PersonChip, {
    name: "Samuel Alderwick",
    years: "1812\u20131878",
    line: "paternal",
    sex: "M"
  }), /*#__PURE__*/React.createElement(PersonChip, {
    name: "Hannah Cobbold",
    years: "1818\u20131889",
    line: "maternal",
    sex: "F",
    grade: "probable"
  }))), /*#__PURE__*/React.createElement(Card, {
    title: "Children",
    subtitle: "6 recorded, 1 inferred",
    padding: "var(--space-6)"
  }, /*#__PURE__*/React.createElement("div", {
    style: {
      display: 'flex',
      flexDirection: 'column',
      gap: 'var(--space-5)'
    }
  }, /*#__PURE__*/React.createElement(PersonChip, {
    size: "sm",
    name: "Mary Ann Alderwick",
    years: "1873\u20131873",
    line: "paternal",
    sex: "F",
    grade: "possible"
  }), /*#__PURE__*/React.createElement(PersonChip, {
    size: "sm",
    name: "Samuel Alderwick",
    years: "1875\u20131948",
    line: "paternal",
    sex: "M"
  }), /*#__PURE__*/React.createElement(PersonChip, {
    size: "sm",
    name: "Ellen Alderwick",
    years: "1878\u20131961",
    line: "paternal",
    sex: "F"
  }), /*#__PURE__*/React.createElement(PersonChip, {
    size: "sm",
    name: "Unnamed infant",
    years: "1879",
    line: "inferred",
    grade: "disputed"
  }))))), /*#__PURE__*/React.createElement("aside", {
    style: {
      width: 'var(--width-inspector)',
      flex: 'none',
      borderLeft: '1px solid var(--border-subtle)',
      background: 'var(--surface-sunken)',
      padding: 'var(--space-8)',
      overflow: 'auto',
      display: 'flex',
      flexDirection: 'column',
      gap: 'var(--space-8)'
    }
  }, /*#__PURE__*/React.createElement("div", null, /*#__PURE__*/React.createElement("div", {
    className: "eyebrow"
  }, "Identity confidence"), /*#__PURE__*/React.createElement("div", {
    style: {
      display: 'flex',
      alignItems: 'baseline',
      gap: 8,
      marginTop: 6
    }
  }, /*#__PURE__*/React.createElement("span", {
    style: {
      fontFamily: 'var(--font-display)',
      fontSize: 'var(--text-display-3)',
      color: 'var(--text-display)'
    }
  }, "Probable"), /*#__PURE__*/React.createElement("span", {
    style: {
      fontFamily: 'var(--font-mono)',
      fontSize: 'var(--text-caption)',
      color: 'var(--text-muted)'
    }
  }, "31 sources")), /*#__PURE__*/React.createElement("p", {
    style: {
      margin: '8px 0 0',
      fontSize: 'var(--text-caption)',
      color: 'var(--text-muted)',
      lineHeight: 'var(--lh-relaxed)'
    }
  }, "Two Thomas Alderwicks appear in Great Yarmouth between 1845 and 1855. Separation rests on the 1871 marriage entry.")), /*#__PURE__*/React.createElement("div", null, /*#__PURE__*/React.createElement("div", {
    className: "eyebrow"
  }, "Open questions"), /*#__PURE__*/React.createElement("div", {
    style: {
      display: 'flex',
      flexDirection: 'column',
      gap: 'var(--space-5)',
      marginTop: 'var(--space-5)'
    }
  }, [['Which Thomas married Ellen Marsh?', 'disputed'], ['Death: 1912 Norwich or 1914 Yarmouth?', 'possible'], ['Naval service record reference', 'undocumented']].map(([q, g]) => /*#__PURE__*/React.createElement("div", {
    key: q,
    style: {
      display: 'flex',
      gap: 'var(--space-5)',
      alignItems: 'flex-start',
      padding: 'var(--space-5)',
      background: 'var(--surface-card)',
      border: '1px solid var(--border-subtle)',
      borderRadius: 'var(--radius-sm)'
    }
  }, /*#__PURE__*/React.createElement(EvidenceBadge, {
    grade: g,
    showLabel: false
  }), /*#__PURE__*/React.createElement("span", {
    style: {
      fontSize: 'var(--text-caption)',
      color: 'var(--text-primary)',
      lineHeight: 1.5
    }
  }, q))))), /*#__PURE__*/React.createElement("div", null, /*#__PURE__*/React.createElement("div", {
    className: "eyebrow"
  }, "Places"), /*#__PURE__*/React.createElement("div", {
    style: {
      display: 'flex',
      flexWrap: 'wrap',
      gap: 'var(--space-4)',
      marginTop: 'var(--space-5)'
    }
  }, /*#__PURE__*/React.createElement(Tag, {
    color: "var(--rec-birth)"
  }, "Great Yarmouth"), /*#__PURE__*/React.createElement(Tag, {
    color: "var(--rec-migration)"
  }, "Grimsby"), /*#__PURE__*/React.createElement(Tag, {
    color: "var(--rec-death)"
  }, "Norwich"))), /*#__PURE__*/React.createElement(Button, {
    variant: "secondary",
    iconLeft: "notebook-pen",
    fullWidth: true
  }, "Add log entry")));
}
Object.assign(window, {
  PersonView
});
})(); } catch (e) { __ds_ns.__errors.push({ path: "ui_kits/app/PersonView.jsx", error: String((e && e.message) || e) }); }

// ui_kits/app/Shell.jsx
try { (() => {
const {
  Icon,
  IconButton,
  Button,
  Badge
} = window.ProvenenciaDesignSystem_0f6c1f;
const {
  SidebarNav,
  Breadcrumbs
} = window.ProvenenciaDesignSystem_0f6c1f;
function TopBar({
  theme,
  onTheme,
  onSearch,
  query,
  setQuery
}) {
  return /*#__PURE__*/React.createElement("header", {
    style: {
      display: 'flex',
      alignItems: 'center',
      gap: 'var(--space-7)',
      height: 56,
      padding: '0 var(--space-8)',
      background: 'var(--surface-card)',
      borderBottom: '1px solid var(--border-subtle)',
      flex: 'none'
    }
  }, /*#__PURE__*/React.createElement("a", {
    href: "#",
    style: {
      display: 'flex',
      alignItems: 'center',
      gap: 9,
      textDecoration: 'none'
    }
  }, /*#__PURE__*/React.createElement("img", {
    src: theme === 'dark' ? '../../assets/logo-mark-inverse.svg' : '../../assets/logo-mark.svg',
    width: "26",
    height: "26",
    alt: ""
  }), /*#__PURE__*/React.createElement("span", {
    style: {
      fontFamily: 'var(--font-display)',
      fontSize: 19,
      fontWeight: 500,
      letterSpacing: '-0.01em',
      color: 'var(--text-display)'
    }
  }, "Provenencia")), /*#__PURE__*/React.createElement("div", {
    style: {
      flex: 1,
      maxWidth: 460,
      position: 'relative'
    }
  }, /*#__PURE__*/React.createElement("form", {
    onSubmit: e => {
      e.preventDefault();
      onSearch();
    },
    style: {
      display: 'flex',
      alignItems: 'center',
      gap: 7,
      height: 'var(--control-h-md)',
      padding: '0 10px',
      background: 'var(--surface-raised)',
      border: '1px solid var(--border-default)',
      borderRadius: 'var(--radius-sm)',
      boxShadow: 'var(--shadow-inset)'
    }
  }, /*#__PURE__*/React.createElement("span", {
    style: {
      color: 'var(--text-faint)',
      display: 'flex'
    }
  }, /*#__PURE__*/React.createElement(Icon, {
    name: "search",
    size: 14
  })), /*#__PURE__*/React.createElement("input", {
    value: query,
    onChange: e => setQuery(e.target.value),
    placeholder: "Search 2.4M records, people and sources",
    style: {
      flex: 1,
      border: 0,
      outline: 'none',
      background: 'transparent',
      color: 'var(--text-primary)',
      fontFamily: 'var(--font-body)',
      fontSize: 'var(--text-body-sm)'
    }
  }), /*#__PURE__*/React.createElement("kbd", {
    style: {
      fontFamily: 'var(--font-mono)',
      fontSize: 10,
      color: 'var(--text-faint)',
      border: '1px solid var(--border-subtle)',
      borderRadius: 3,
      padding: '1px 4px'
    }
  }, "\u2318K"))), /*#__PURE__*/React.createElement("div", {
    style: {
      display: 'flex',
      alignItems: 'center',
      gap: 'var(--space-4)',
      marginLeft: 'auto'
    }
  }, /*#__PURE__*/React.createElement(Badge, {
    tone: "accent",
    icon: "lock"
  }, "Private tree"), /*#__PURE__*/React.createElement(IconButton, {
    icon: theme === 'dark' ? 'sun' : 'moon',
    label: "Toggle theme",
    onClick: onTheme
  }), /*#__PURE__*/React.createElement(IconButton, {
    icon: "bell",
    label: "Notifications"
  }), /*#__PURE__*/React.createElement("span", {
    style: {
      width: 28,
      height: 28,
      borderRadius: 'var(--radius-xs)',
      background: 'var(--iron-700)',
      color: 'var(--paper-0)',
      display: 'grid',
      placeItems: 'center',
      fontFamily: 'var(--font-display)',
      fontSize: 13
    }
  }, "HA")));
}
function Sidebar({
  view,
  onView
}) {
  return /*#__PURE__*/React.createElement("aside", {
    style: {
      width: 'var(--width-sidebar)',
      flex: 'none',
      background: 'var(--surface-sunken)',
      borderRight: '1px solid var(--border-subtle)',
      padding: 'var(--space-7) var(--space-5)',
      display: 'flex',
      flexDirection: 'column',
      gap: 'var(--space-8)',
      overflow: 'auto'
    }
  }, /*#__PURE__*/React.createElement("div", {
    style: {
      padding: '0 10px'
    }
  }, /*#__PURE__*/React.createElement("div", {
    className: "eyebrow"
  }, "Workspace"), /*#__PURE__*/React.createElement("div", {
    style: {
      display: 'flex',
      alignItems: 'center',
      gap: 6,
      marginTop: 5
    }
  }, /*#__PURE__*/React.createElement("span", {
    style: {
      fontFamily: 'var(--font-display)',
      fontSize: 'var(--text-h3)',
      color: 'var(--text-display)'
    }
  }, "Alderwick line"), /*#__PURE__*/React.createElement(Icon, {
    name: "chevron-down",
    size: 14,
    style: {
      color: 'var(--text-faint)'
    }
  }))), /*#__PURE__*/React.createElement(SidebarNav, {
    value: view,
    onChange: onView,
    groups: [{
      label: 'Research',
      items: [{
        id: 'dashboard',
        label: 'Overview',
        icon: 'layout-dashboard'
      }, {
        id: 'people',
        label: 'People',
        icon: 'users',
        count: 412
      }, {
        id: 'tree',
        label: 'Pedigree',
        icon: 'git-branch'
      }, {
        id: 'sources',
        label: 'Sources',
        icon: 'scroll-text',
        count: 1284
      }, {
        id: 'log',
        label: 'Research log',
        icon: 'notebook-pen',
        count: 26
      }]
    }, {
      label: 'Collections',
      items: [{
        id: 'census',
        label: 'Census',
        icon: 'table-2',
        count: 212
      }, {
        id: 'vitals',
        label: 'Civil registration',
        icon: 'file-text',
        count: 341
      }, {
        id: 'dna',
        label: 'DNA matches',
        icon: 'dna',
        count: 37
      }, {
        id: 'maps',
        label: 'Places',
        icon: 'map-pin',
        count: 64
      }]
    }]
  }), /*#__PURE__*/React.createElement("div", {
    style: {
      marginTop: 'auto',
      padding: 'var(--space-6)',
      background: 'var(--surface-card)',
      border: '1px solid var(--border-subtle)',
      borderRadius: 'var(--radius-sm)'
    }
  }, /*#__PURE__*/React.createElement("div", {
    className: "eyebrow"
  }, "Evidence health"), /*#__PURE__*/React.createElement("div", {
    style: {
      display: 'flex',
      height: 6,
      borderRadius: 2,
      overflow: 'hidden',
      marginTop: 9
    }
  }, /*#__PURE__*/React.createElement("span", {
    style: {
      flex: 52,
      background: 'var(--evidence-proven)'
    }
  }), /*#__PURE__*/React.createElement("span", {
    style: {
      flex: 23,
      background: 'var(--evidence-probable)'
    }
  }), /*#__PURE__*/React.createElement("span", {
    style: {
      flex: 14,
      background: 'var(--evidence-possible)'
    }
  }), /*#__PURE__*/React.createElement("span", {
    style: {
      flex: 5,
      background: 'var(--evidence-disputed)'
    }
  }), /*#__PURE__*/React.createElement("span", {
    style: {
      flex: 6,
      background: 'var(--evidence-undocumented)'
    }
  })), /*#__PURE__*/React.createElement("p", {
    style: {
      margin: '8px 0 0',
      fontSize: 'var(--text-micro)',
      color: 'var(--text-muted)',
      lineHeight: 1.6
    }
  }, "52% of 1,904 facts are proven. 94 conflicts await review.")));
}
function PageHeader({
  crumbs,
  title,
  meta,
  actions,
  tabs
}) {
  return /*#__PURE__*/React.createElement("div", {
    style: {
      padding: 'var(--space-8) var(--gutter-page) 0'
    }
  }, crumbs && /*#__PURE__*/React.createElement(Breadcrumbs, {
    items: crumbs
  }), /*#__PURE__*/React.createElement("div", {
    style: {
      display: 'flex',
      alignItems: 'flex-end',
      gap: 'var(--space-8)',
      marginTop: 'var(--space-4)'
    }
  }, /*#__PURE__*/React.createElement("div", null, /*#__PURE__*/React.createElement("h1", {
    style: {
      fontFamily: 'var(--font-display)',
      fontSize: 'var(--text-display-3)',
      letterSpacing: 'var(--tracking-display)',
      fontWeight: 600,
      color: 'var(--text-display)',
      margin: 0
    }
  }, title), meta && /*#__PURE__*/React.createElement("div", {
    style: {
      display: 'flex',
      alignItems: 'center',
      gap: 'var(--space-5)',
      marginTop: 6
    }
  }, meta)), /*#__PURE__*/React.createElement("div", {
    style: {
      marginLeft: 'auto',
      display: 'flex',
      gap: 'var(--space-4)'
    }
  }, actions)), tabs && /*#__PURE__*/React.createElement("div", {
    style: {
      marginTop: 'var(--space-7)'
    }
  }, tabs));
}
Object.assign(window, {
  TopBar,
  Sidebar,
  PageHeader
});
})(); } catch (e) { __ds_ns.__errors.push({ path: "ui_kits/app/Shell.jsx", error: String((e && e.message) || e) }); }

// ui_kits/app/SourcesLibrary.jsx
try { (() => {
const {
  Button,
  IconButton,
  Badge,
  Icon,
  Input,
  Select,
  Checkbox,
  EvidenceBadge,
  Tag,
  EmptyState
} = window.ProvenenciaDesignSystem_0f6c1f;
const ROWS = [{
  t: '1881 census of England — Great Yarmouth, ED 12',
  rep: 'The National Archives',
  type: 'census',
  people: 6,
  grade: 'proven',
  date: '14 Mar 2026',
  img: true
}, {
  t: 'Marriages and banns, Great Yarmouth St Nicholas, 1871',
  rep: 'Norfolk Record Office',
  type: 'marriage',
  people: 2,
  grade: 'probable',
  date: '12 Mar 2026',
  img: true
}, {
  t: 'GRO death index, Norwich district, Q2 1912',
  rep: 'General Register Office',
  type: 'probate',
  people: 1,
  grade: 'disputed',
  date: '09 Mar 2026',
  img: false
}, {
  t: 'Crew list, SS Iceni, Grimsby 1883',
  rep: 'Maritime History Archive',
  type: 'migration',
  people: 1,
  grade: 'possible',
  date: '02 Mar 2026',
  img: true
}, {
  t: 'Royal Naval Reserve service registers, Yarmouth division',
  rep: 'The National Archives (ADM 240)',
  type: 'military',
  people: 1,
  grade: 'possible',
  date: '27 Feb 2026',
  img: false
}, {
  t: 'Autosomal match report — 42 cM shared, 3rd cousin range',
  rep: 'Provenencia DNA',
  type: 'dna',
  people: 4,
  grade: 'probable',
  date: '21 Feb 2026',
  img: false
}];
function SourcesLibrary({
  onAttach
}) {
  const [sel, setSel] = React.useState([1]);
  const toggle = i => setSel(s => s.includes(i) ? s.filter(x => x !== i) : [...s, i]);
  return /*#__PURE__*/React.createElement("div", {
    style: {
      flex: 1,
      minHeight: 0,
      display: 'flex',
      flexDirection: 'column'
    }
  }, /*#__PURE__*/React.createElement("div", {
    style: {
      display: 'flex',
      alignItems: 'center',
      gap: 'var(--space-5)',
      padding: 'var(--space-6) var(--gutter-page)',
      borderBottom: '1px solid var(--border-subtle)',
      background: 'var(--surface-card)'
    }
  }, /*#__PURE__*/React.createElement("div", {
    style: {
      width: 260
    }
  }, /*#__PURE__*/React.createElement(Input, {
    iconLeft: "search",
    placeholder: "Filter 1,284 sources",
    size: "sm"
  })), /*#__PURE__*/React.createElement("div", {
    style: {
      width: 170
    }
  }, /*#__PURE__*/React.createElement(Select, {
    size: "sm",
    options: ['All repositories', 'The National Archives', 'Norfolk Record Office', 'GRO']
  })), /*#__PURE__*/React.createElement("div", {
    style: {
      width: 150
    }
  }, /*#__PURE__*/React.createElement(Select, {
    size: "sm",
    options: ['All grades', 'Proven', 'Probable', 'Possible', 'Disputed']
  })), /*#__PURE__*/React.createElement("div", {
    style: {
      marginLeft: 'auto',
      display: 'flex',
      gap: 'var(--space-4)'
    }
  }, /*#__PURE__*/React.createElement(Button, {
    size: "sm",
    variant: "ghost",
    iconLeft: "upload"
  }, "Import"), /*#__PURE__*/React.createElement(Button, {
    size: "sm",
    variant: "primary",
    iconLeft: "plus",
    onClick: onAttach
  }, "New source"))), /*#__PURE__*/React.createElement("div", {
    style: {
      flex: 1,
      overflow: 'auto'
    }
  }, /*#__PURE__*/React.createElement("table", {
    style: {
      width: '100%',
      borderCollapse: 'collapse',
      fontFamily: 'var(--font-body)'
    }
  }, /*#__PURE__*/React.createElement("thead", null, /*#__PURE__*/React.createElement("tr", {
    style: {
      background: 'var(--surface-sunken)'
    }
  }, ['', 'Source', 'Repository', 'Type', 'Cites', 'Grade', 'Accessed', ''].map((h, i) => /*#__PURE__*/React.createElement("th", {
    key: i,
    style: {
      textAlign: i > 3 && i < 7 ? 'right' : 'left',
      fontSize: 'var(--text-micro)',
      letterSpacing: 'var(--tracking-caps)',
      textTransform: 'uppercase',
      color: 'var(--text-faint)',
      fontWeight: 600,
      padding: 'var(--space-5) var(--space-6)',
      borderBottom: '1px solid var(--border-default)',
      position: 'sticky',
      top: 0,
      background: 'var(--surface-sunken)'
    }
  }, h)))), /*#__PURE__*/React.createElement("tbody", null, ROWS.map((r, i) => /*#__PURE__*/React.createElement("tr", {
    key: i,
    style: {
      background: sel.includes(i) ? 'var(--surface-selected)' : 'transparent',
      borderBottom: '1px solid var(--border-subtle)'
    }
  }, /*#__PURE__*/React.createElement("td", {
    style: {
      padding: 'var(--space-5) var(--space-6)',
      width: 34
    }
  }, /*#__PURE__*/React.createElement(Checkbox, {
    checked: sel.includes(i),
    onChange: () => toggle(i)
  })), /*#__PURE__*/React.createElement("td", {
    style: {
      padding: 'var(--space-5) var(--space-6)'
    }
  }, /*#__PURE__*/React.createElement("div", {
    style: {
      display: 'flex',
      alignItems: 'center',
      gap: 'var(--space-5)'
    }
  }, /*#__PURE__*/React.createElement("span", {
    style: {
      width: 26,
      height: 32,
      flex: 'none',
      borderRadius: 2,
      border: '1px solid var(--border-default)',
      background: 'var(--surface-inset)',
      display: 'grid',
      placeItems: 'center',
      color: 'var(--text-faint)'
    }
  }, /*#__PURE__*/React.createElement(Icon, {
    name: r.img ? 'image' : 'file-text',
    size: 12
  })), /*#__PURE__*/React.createElement("span", {
    style: {
      fontSize: 'var(--text-body-sm)',
      color: 'var(--text-primary)'
    }
  }, r.t))), /*#__PURE__*/React.createElement("td", {
    style: {
      padding: 'var(--space-5) var(--space-6)',
      fontSize: 'var(--text-caption)',
      color: 'var(--text-muted)'
    }
  }, r.rep), /*#__PURE__*/React.createElement("td", {
    style: {
      padding: 'var(--space-5) var(--space-6)'
    }
  }, /*#__PURE__*/React.createElement(Tag, {
    color: 'var(--rec-' + r.type + ')'
  }, r.type)), /*#__PURE__*/React.createElement("td", {
    style: {
      padding: 'var(--space-5) var(--space-6)',
      textAlign: 'right',
      fontFamily: 'var(--font-mono)',
      fontSize: 'var(--text-caption)',
      color: 'var(--text-secondary)'
    }
  }, r.people), /*#__PURE__*/React.createElement("td", {
    style: {
      padding: 'var(--space-5) var(--space-6)',
      textAlign: 'right'
    }
  }, /*#__PURE__*/React.createElement(EvidenceBadge, {
    grade: r.grade
  })), /*#__PURE__*/React.createElement("td", {
    style: {
      padding: 'var(--space-5) var(--space-6)',
      textAlign: 'right',
      fontFamily: 'var(--font-mono)',
      fontSize: 'var(--text-micro)',
      color: 'var(--text-faint)'
    }
  }, r.date), /*#__PURE__*/React.createElement("td", {
    style: {
      padding: 'var(--space-5) var(--space-6)',
      width: 36
    }
  }, /*#__PURE__*/React.createElement(IconButton, {
    icon: "more-horizontal",
    label: "Row actions",
    size: "sm"
  })))))), sel.length > 0 && /*#__PURE__*/React.createElement("div", {
    style: {
      position: 'sticky',
      bottom: 0,
      display: 'flex',
      alignItems: 'center',
      gap: 'var(--space-6)',
      padding: 'var(--space-5) var(--gutter-page)',
      background: 'var(--surface-card)',
      borderTop: '1px solid var(--border-default)',
      boxShadow: 'var(--shadow-lg)'
    }
  }, /*#__PURE__*/React.createElement("span", {
    style: {
      fontSize: 'var(--text-caption)',
      color: 'var(--text-secondary)'
    }
  }, sel.length, " selected"), /*#__PURE__*/React.createElement(Button, {
    size: "sm",
    variant: "secondary",
    iconLeft: "paperclip",
    onClick: onAttach
  }, "Attach to person"), /*#__PURE__*/React.createElement(Button, {
    size: "sm",
    variant: "ghost",
    iconLeft: "quote"
  }, "Copy citations"), /*#__PURE__*/React.createElement(Button, {
    size: "sm",
    variant: "ghost",
    iconLeft: "trash-2"
  }, "Remove"))));
}
Object.assign(window, {
  SourcesLibrary
});
})(); } catch (e) { __ds_ns.__errors.push({ path: "ui_kits/app/SourcesLibrary.jsx", error: String((e && e.message) || e) }); }

// ui_kits/app/TreeView.jsx
try { (() => {
function _extends() { return _extends = Object.assign ? Object.assign.bind() : function (n) { for (var e = 1; e < arguments.length; e++) { var t = arguments[e]; for (var r in t) ({}).hasOwnProperty.call(t, r) && (n[r] = t[r]); } return n; }, _extends.apply(null, arguments); }
const {
  PersonChip,
  IconButton,
  Button,
  Badge,
  Icon,
  Switch
} = window.ProvenenciaDesignSystem_0f6c1f;
function Node({
  p,
  onOpen
}) {
  return /*#__PURE__*/React.createElement(PersonChip, _extends({}, p, {
    onClick: onOpen
  }));
}
function TreeView({
  onOpenPerson
}) {
  const [inferred, setInferred] = React.useState(true);
  const col = {
    display: 'flex',
    flexDirection: 'column',
    justifyContent: 'space-around',
    gap: 'var(--space-6)'
  };
  const rule = h => ({
    width: 26,
    borderTop: '1px solid var(--line-paternal)',
    height: h
  });
  return /*#__PURE__*/React.createElement("div", {
    style: {
      flex: 1,
      minHeight: 0,
      display: 'flex',
      flexDirection: 'column'
    }
  }, /*#__PURE__*/React.createElement("div", {
    style: {
      display: 'flex',
      alignItems: 'center',
      gap: 'var(--space-6)',
      padding: 'var(--space-6) var(--gutter-page)',
      borderBottom: '1px solid var(--border-subtle)',
      background: 'var(--surface-card)'
    }
  }, /*#__PURE__*/React.createElement("div", {
    style: {
      display: 'flex',
      gap: 2
    }
  }, /*#__PURE__*/React.createElement(IconButton, {
    icon: "git-branch",
    label: "Pedigree",
    active: true,
    variant: "outline"
  }), /*#__PURE__*/React.createElement(IconButton, {
    icon: "network",
    label: "Descendants",
    variant: "outline"
  }), /*#__PURE__*/React.createElement(IconButton, {
    icon: "table-2",
    label: "Table",
    variant: "outline"
  })), /*#__PURE__*/React.createElement(Switch, {
    checked: inferred,
    onChange: () => setInferred(!inferred),
    label: "Show inferred relationships"
  }), /*#__PURE__*/React.createElement("div", {
    style: {
      marginLeft: 'auto',
      display: 'flex',
      alignItems: 'center',
      gap: 'var(--space-6)'
    }
  }, /*#__PURE__*/React.createElement("span", {
    style: {
      display: 'inline-flex',
      alignItems: 'center',
      gap: 6,
      fontSize: 'var(--text-micro)',
      color: 'var(--text-muted)'
    }
  }, /*#__PURE__*/React.createElement("span", {
    style: {
      width: 14,
      borderTop: '2px solid var(--line-paternal)'
    }
  }), "Paternal", /*#__PURE__*/React.createElement("span", {
    style: {
      width: 14,
      borderTop: '2px solid var(--line-maternal)',
      marginLeft: 8
    }
  }), "Maternal", /*#__PURE__*/React.createElement("span", {
    style: {
      width: 14,
      borderTop: '2px dashed var(--line-inferred)',
      marginLeft: 8
    }
  }), "Inferred"), /*#__PURE__*/React.createElement(Button, {
    size: "sm",
    variant: "secondary",
    iconLeft: "download"
  }, "Export GEDCOM"))), /*#__PURE__*/React.createElement("div", {
    style: {
      flex: 1,
      overflow: 'auto',
      padding: 'var(--space-11) var(--gutter-page)',
      background: 'var(--surface-page)'
    }
  }, /*#__PURE__*/React.createElement("div", {
    style: {
      display: 'flex',
      alignItems: 'center',
      gap: 0
    }
  }, /*#__PURE__*/React.createElement("div", {
    style: col
  }, /*#__PURE__*/React.createElement(Node, {
    onOpen: onOpenPerson,
    p: {
      name: 'Thomas Alderwick',
      years: '1847–1912',
      line: 'paternal',
      sex: 'M',
      selected: true,
      grade: 'disputed'
    }
  })), /*#__PURE__*/React.createElement("div", {
    style: rule(1)
  }), /*#__PURE__*/React.createElement("div", {
    style: {
      borderLeft: '1px solid var(--border-default)',
      ...col,
      height: 280,
      paddingLeft: 0
    }
  }, /*#__PURE__*/React.createElement("div", {
    style: {
      display: 'flex',
      alignItems: 'center'
    }
  }, /*#__PURE__*/React.createElement("div", {
    style: rule(1)
  }), /*#__PURE__*/React.createElement(Node, {
    onOpen: onOpenPerson,
    p: {
      name: 'Samuel Alderwick',
      years: '1812–1878',
      line: 'paternal',
      sex: 'M'
    }
  })), /*#__PURE__*/React.createElement("div", {
    style: {
      display: 'flex',
      alignItems: 'center'
    }
  }, /*#__PURE__*/React.createElement("div", {
    style: {
      ...rule(1),
      borderColor: 'var(--line-maternal)'
    }
  }), /*#__PURE__*/React.createElement(Node, {
    onOpen: onOpenPerson,
    p: {
      name: 'Hannah Cobbold',
      years: '1818–1889',
      line: 'maternal',
      sex: 'F',
      grade: 'probable'
    }
  }))), /*#__PURE__*/React.createElement("div", {
    style: {
      borderLeft: '1px solid var(--border-subtle)',
      display: 'flex',
      flexDirection: 'column',
      justifyContent: 'space-around',
      height: 400,
      marginLeft: 26
    }
  }, [{
    name: 'Josiah Alderwick',
    years: '1780–1841',
    line: 'paternal',
    sex: 'M',
    size: 'sm'
  }, {
    name: 'Susanna Pye',
    years: '1786–1859',
    line: 'maternal',
    sex: 'F',
    size: 'sm'
  }, {
    name: 'William Cobbold',
    years: '1789–1852',
    line: 'paternal',
    sex: 'M',
    size: 'sm',
    grade: 'possible'
  }, {
    name: 'Unknown mother',
    years: '—',
    line: 'inferred',
    sex: 'U',
    size: 'sm',
    grade: 'undocumented'
  }].filter(p => inferred || p.line !== 'inferred').map((p, i) => /*#__PURE__*/React.createElement("div", {
    key: i,
    style: {
      display: 'flex',
      alignItems: 'center'
    }
  }, /*#__PURE__*/React.createElement("div", {
    style: {
      width: 26,
      borderTop: '1px ' + (p.line === 'inferred' ? 'dashed var(--line-inferred)' : 'solid var(--line-paternal)')
    }
  }), /*#__PURE__*/React.createElement(Node, {
    onOpen: onOpenPerson,
    p: p
  })))))));
}
Object.assign(window, {
  TreeView
});
})(); } catch (e) { __ds_ns.__errors.push({ path: "ui_kits/app/TreeView.jsx", error: String((e && e.message) || e) }); }

// ui_kits/site/Site.jsx
try { (() => {
const {
  Button,
  Badge,
  Icon,
  Card,
  EvidenceBadge,
  SourceCitation,
  PersonChip,
  Tag,
  FactRow,
  IconButton,
  Input
} = window.ProvenenciaDesignSystem_0f6c1f;
function Nav({
  onTheme,
  theme
}) {
  return /*#__PURE__*/React.createElement("header", {
    style: {
      position: 'sticky',
      top: 0,
      zIndex: 20,
      display: 'flex',
      alignItems: 'center',
      gap: 'var(--space-9)',
      padding: '0 var(--gutter-page-wide)',
      height: 66,
      background: 'color-mix(in oklab, var(--surface-page) 88%, transparent)',
      backdropFilter: 'blur(8px)',
      borderBottom: '1px solid var(--border-subtle)'
    }
  }, /*#__PURE__*/React.createElement("a", {
    href: "#",
    style: {
      display: 'flex',
      alignItems: 'center',
      gap: 10,
      textDecoration: 'none'
    }
  }, /*#__PURE__*/React.createElement("img", {
    src: theme === 'dark' ? '../../assets/logo-mark-inverse.svg' : '../../assets/logo-mark.svg',
    width: "30",
    height: "30",
    alt: ""
  }), /*#__PURE__*/React.createElement("span", {
    style: {
      fontFamily: 'var(--font-display)',
      fontSize: 22,
      fontWeight: 500,
      color: 'var(--text-display)'
    }
  }, "Provenencia")), /*#__PURE__*/React.createElement("nav", {
    style: {
      display: 'flex',
      gap: 'var(--space-8)'
    }
  }, ['How it works', 'Records', 'Method', 'Pricing'].map(l => /*#__PURE__*/React.createElement("a", {
    key: l,
    href: "#",
    style: {
      fontSize: 'var(--text-body-sm)',
      color: 'var(--text-secondary)',
      textDecoration: 'none'
    }
  }, l))), /*#__PURE__*/React.createElement("div", {
    style: {
      marginLeft: 'auto',
      display: 'flex',
      alignItems: 'center',
      gap: 'var(--space-5)'
    }
  }, /*#__PURE__*/React.createElement(IconButton, {
    icon: theme === 'dark' ? 'sun' : 'moon',
    label: "Toggle theme",
    onClick: onTheme
  }), /*#__PURE__*/React.createElement(Button, {
    variant: "ghost",
    size: "sm"
  }, "Sign in"), /*#__PURE__*/React.createElement(Button, {
    variant: "primary",
    size: "sm",
    iconRight: "arrow-right"
  }, "Start a tree")));
}
function Hero() {
  return /*#__PURE__*/React.createElement("section", {
    style: {
      display: 'grid',
      gridTemplateColumns: '1.05fr .95fr',
      gap: 'var(--space-13)',
      alignItems: 'center',
      padding: 'var(--space-14) var(--gutter-page-wide) var(--space-13)',
      maxWidth: 1320,
      margin: '0 auto'
    }
  }, /*#__PURE__*/React.createElement("div", null, /*#__PURE__*/React.createElement("div", {
    className: "eyebrow"
  }, "Evidence-based genealogy"), /*#__PURE__*/React.createElement("h1", {
    style: {
      fontFamily: 'var(--font-display)',
      fontSize: 'var(--text-display-1)',
      lineHeight: 1.04,
      letterSpacing: 'var(--tracking-display)',
      fontWeight: 500,
      color: 'var(--text-display)',
      margin: 'var(--space-6) 0 0',
      maxWidth: '18ch'
    }
  }, "A family history that can be checked."), /*#__PURE__*/React.createElement("p", {
    style: {
      fontSize: 19,
      lineHeight: 'var(--lh-relaxed)',
      color: 'var(--text-secondary)',
      maxWidth: '46ch',
      margin: 'var(--space-7) 0 0'
    }
  }, "Provenencia keeps the record beside the conclusion. Every name, date and relationship carries its citation, its evidence grade, and the reasoning that got you there."), /*#__PURE__*/React.createElement("div", {
    style: {
      display: 'flex',
      gap: 'var(--space-5)',
      marginTop: 'var(--space-9)'
    }
  }, /*#__PURE__*/React.createElement(Button, {
    variant: "primary",
    size: "lg",
    iconRight: "arrow-right"
  }, "Start a tree"), /*#__PURE__*/React.createElement(Button, {
    variant: "secondary",
    size: "lg",
    iconLeft: "play"
  }, "See the method")), /*#__PURE__*/React.createElement("div", {
    style: {
      display: 'flex',
      alignItems: 'center',
      gap: 'var(--space-7)',
      marginTop: 'var(--space-9)',
      fontSize: 'var(--text-caption)',
      color: 'var(--text-muted)'
    }
  }, /*#__PURE__*/React.createElement("span", {
    style: {
      display: 'inline-flex',
      alignItems: 'center',
      gap: 6
    }
  }, /*#__PURE__*/React.createElement(Icon, {
    name: "library",
    size: 14
  }), "61 repositories indexed"), /*#__PURE__*/React.createElement("span", {
    style: {
      display: 'inline-flex',
      alignItems: 'center',
      gap: 6
    }
  }, /*#__PURE__*/React.createElement(Icon, {
    name: "scroll-text",
    size: 14
  }), "2.4M records"), /*#__PURE__*/React.createElement("span", {
    style: {
      display: 'inline-flex',
      alignItems: 'center',
      gap: 6
    }
  }, /*#__PURE__*/React.createElement(Icon, {
    name: "file-down",
    size: 14
  }), "GEDCOM in and out"))), /*#__PURE__*/React.createElement("div", {
    style: {
      background: 'var(--surface-card)',
      border: '1px solid var(--border-subtle)',
      borderRadius: 'var(--radius-lg)',
      boxShadow: 'var(--shadow-lg)',
      overflow: 'hidden'
    }
  }, /*#__PURE__*/React.createElement("div", {
    style: {
      display: 'flex',
      alignItems: 'center',
      gap: 'var(--space-5)',
      padding: 'var(--space-6) var(--space-7)',
      borderBottom: '1px solid var(--border-subtle)',
      background: 'var(--surface-sunken)'
    }
  }, /*#__PURE__*/React.createElement(PersonChip, {
    name: "Thomas Alderwick",
    years: "1847\u20131912",
    line: "paternal",
    sex: "M"
  }), /*#__PURE__*/React.createElement(EvidenceBadge, {
    grade: "probable",
    sources: 31
  })), /*#__PURE__*/React.createElement(FactRow, {
    type: "birth",
    date: "12 Mar 1847",
    value: "Great Yarmouth, Norfolk",
    place: "St Nicholas parish register",
    sources: 3,
    grade: "proven"
  }), /*#__PURE__*/React.createElement(FactRow, {
    type: "census",
    date: "3 Apr 1881",
    value: "12 Row 62, Great Yarmouth",
    place: "RG11/1976, sched. 41",
    sources: 2,
    grade: "proven"
  }), /*#__PURE__*/React.createElement(FactRow, {
    type: "death",
    date: "abt 1912",
    value: "Norwich",
    sources: 1,
    grade: "disputed",
    conflict: "Conflicts with 1891 census age"
  }), /*#__PURE__*/React.createElement("div", {
    style: {
      padding: 'var(--space-6) var(--space-7)',
      background: 'var(--surface-sunken)'
    }
  }, /*#__PURE__*/React.createElement(SourceCitation, {
    style: {
      border: 0,
      background: 'transparent',
      padding: 0
    },
    citation: "1881 census of England, Norfolk, Great Yarmouth, ED 12, sched. 41, Thomas Alderwick household.",
    repository: "The National Archives (RG11/1976)",
    accessed: "accessed 14 Mar 2026",
    quality: "Original \xB7 digital image"
  }))));
}
function Grades() {
  const items = [['proven', 'Proven', 'Two or more independent originals agree.'], ['probable', 'Probable', 'One original, or strong indirect evidence.'], ['possible', 'Possible', 'A plausible match with unresolved questions.'], ['disputed', 'Disputed', 'Sources conflict and the conflict is recorded.'], ['undocumented', 'Undocumented', 'Asserted, with nothing behind it yet.']];
  return /*#__PURE__*/React.createElement("section", {
    style: {
      padding: 'var(--space-13) var(--gutter-page-wide)',
      background: 'var(--surface-sunken)',
      borderTop: '1px solid var(--border-subtle)',
      borderBottom: '1px solid var(--border-subtle)'
    }
  }, /*#__PURE__*/React.createElement("div", {
    style: {
      maxWidth: 1320,
      margin: '0 auto'
    }
  }, /*#__PURE__*/React.createElement("div", {
    className: "eyebrow"
  }, "Five grades, one vocabulary"), /*#__PURE__*/React.createElement("h2", {
    style: {
      fontFamily: 'var(--font-display)',
      fontSize: 'var(--text-display-3)',
      letterSpacing: 'var(--tracking-display)',
      color: 'var(--text-display)',
      margin: 'var(--space-5) 0 var(--space-9)',
      maxWidth: '26ch'
    }
  }, "Confidence is stated, not implied."), /*#__PURE__*/React.createElement("div", {
    style: {
      display: 'grid',
      gridTemplateColumns: 'repeat(5,1fr)',
      gap: 'var(--space-6)'
    }
  }, items.map(([g, t, d]) => /*#__PURE__*/React.createElement("div", {
    key: g,
    style: {
      padding: 'var(--space-7)',
      background: 'var(--surface-card)',
      border: '1px solid var(--border-subtle)',
      borderRadius: 'var(--radius-md)'
    }
  }, /*#__PURE__*/React.createElement(EvidenceBadge, {
    grade: g
  }), /*#__PURE__*/React.createElement("div", {
    style: {
      fontFamily: 'var(--font-display)',
      fontSize: 'var(--text-h3)',
      color: 'var(--text-display)',
      marginTop: 'var(--space-6)'
    }
  }, t), /*#__PURE__*/React.createElement("p", {
    style: {
      margin: '6px 0 0',
      fontSize: 'var(--text-caption)',
      lineHeight: 'var(--lh-relaxed)',
      color: 'var(--text-muted)'
    }
  }, d))))));
}
function Features() {
  const feats = [['scroll-text', 'Citation first', 'Add the record, then the conclusion. Provenencia writes the citation in Evidence Explained form and keeps the image beside it.'], ['git-compare-arrows', 'Conflicts kept, not hidden', 'Contradictory ages, dates and places stay on the record as open questions until you resolve them in writing.'], ['notebook-pen', 'A research log that writes itself', 'Every search, order and negative result is logged with its date — so you never repeat a dead end.'], ['table-2', 'Record-type colour coding', 'Census, civil registration, migration, military, probate and DNA each keep their own hue across the whole workspace.'], ['users', 'Collateral lines, properly', 'Trace siblings and neighbours as first-class research subjects, not footnotes to a direct line.'], ['file-down', 'Yours to take', 'Full GEDCOM export with citations intact, plus a plain-text research report.']];
  return /*#__PURE__*/React.createElement("section", {
    style: {
      padding: 'var(--space-13) var(--gutter-page-wide)',
      maxWidth: 1320,
      margin: '0 auto'
    }
  }, /*#__PURE__*/React.createElement("div", {
    style: {
      display: 'grid',
      gridTemplateColumns: 'repeat(3,1fr)',
      gap: 'var(--space-9) var(--space-8)'
    }
  }, feats.map(([ic, t, d]) => /*#__PURE__*/React.createElement("div", {
    key: t
  }, /*#__PURE__*/React.createElement("span", {
    style: {
      display: 'grid',
      placeItems: 'center',
      width: 38,
      height: 38,
      borderRadius: 'var(--radius-sm)',
      background: 'var(--accent-soft)',
      color: 'var(--accent)'
    }
  }, /*#__PURE__*/React.createElement(Icon, {
    name: ic,
    size: 18
  })), /*#__PURE__*/React.createElement("h3", {
    style: {
      fontFamily: 'var(--font-display)',
      fontSize: 'var(--text-h2)',
      color: 'var(--text-display)',
      margin: 'var(--space-6) 0 var(--space-4)'
    }
  }, t), /*#__PURE__*/React.createElement("p", {
    style: {
      margin: 0,
      fontSize: 'var(--text-body-sm)',
      lineHeight: 'var(--lh-relaxed)',
      color: 'var(--text-secondary)',
      maxWidth: '42ch'
    }
  }, d)))));
}
function Closing() {
  return /*#__PURE__*/React.createElement("section", {
    style: {
      padding: 'var(--space-13) var(--gutter-page-wide)',
      background: 'var(--paper-950)',
      color: 'var(--paper-100)'
    }
  }, /*#__PURE__*/React.createElement("div", {
    style: {
      maxWidth: 1320,
      margin: '0 auto',
      display: 'grid',
      gridTemplateColumns: '1.2fr .8fr',
      gap: 'var(--space-12)',
      alignItems: 'center'
    }
  }, /*#__PURE__*/React.createElement("div", null, /*#__PURE__*/React.createElement("img", {
    src: "../../assets/logo-mark-inverse.svg",
    width: "40",
    height: "40",
    alt: ""
  }), /*#__PURE__*/React.createElement("h2", {
    style: {
      fontFamily: 'var(--font-display)',
      fontSize: 'var(--text-display-2)',
      lineHeight: 1.1,
      letterSpacing: 'var(--tracking-display)',
      fontWeight: 500,
      color: 'var(--paper-50)',
      margin: 'var(--space-7) 0 0',
      maxWidth: '22ch'
    }
  }, "Begin with one certificate you can hold."), /*#__PURE__*/React.createElement("p", {
    style: {
      fontSize: 'var(--text-body)',
      lineHeight: 'var(--lh-relaxed)',
      color: 'var(--paper-300)',
      maxWidth: '48ch',
      marginTop: 'var(--space-6)'
    }
  }, "Free while your tree is under 100 people. Import an existing GEDCOM and Provenencia will grade what it finds.")), /*#__PURE__*/React.createElement("div", {
    style: {
      display: 'flex',
      flexDirection: 'column',
      gap: 'var(--space-5)'
    }
  }, /*#__PURE__*/React.createElement(Input, {
    iconLeft: "mail",
    placeholder: "you@example.com",
    style: {
      background: 'var(--paper-900)',
      borderColor: 'var(--paper-700)'
    }
  }), /*#__PURE__*/React.createElement(Button, {
    variant: "primary",
    size: "lg",
    fullWidth: true,
    iconRight: "arrow-right",
    style: {
      background: 'var(--paper-50)',
      borderColor: 'var(--paper-50)',
      color: 'var(--paper-950)'
    }
  }, "Start a tree"), /*#__PURE__*/React.createElement("span", {
    style: {
      fontSize: 'var(--text-micro)',
      color: 'var(--paper-400)'
    }
  }, "No credit card. Export anything you add."))));
}
function Footer() {
  const cols = [['Product', ['How it works', 'Records', 'Method', 'Pricing']], ['Research', ['Citation guide', 'Evidence grades', 'Research logs', 'GEDCOM support']], ['Company', ['About', 'Archive partners', 'Privacy', 'Contact']]];
  return /*#__PURE__*/React.createElement("footer", {
    style: {
      padding: 'var(--space-11) var(--gutter-page-wide)',
      background: 'var(--surface-page)',
      borderTop: '1px solid var(--border-subtle)'
    }
  }, /*#__PURE__*/React.createElement("div", {
    style: {
      maxWidth: 1320,
      margin: '0 auto',
      display: 'grid',
      gridTemplateColumns: '1.4fr repeat(3,1fr)',
      gap: 'var(--space-9)'
    }
  }, /*#__PURE__*/React.createElement("div", null, /*#__PURE__*/React.createElement("img", {
    src: "../../assets/logo-lockup.svg",
    width: "220",
    height: "36",
    alt: "Provenencia"
  }), /*#__PURE__*/React.createElement("p", {
    style: {
      fontSize: 'var(--text-caption)',
      color: 'var(--text-muted)',
      maxWidth: '34ch',
      marginTop: 'var(--space-6)'
    }
  }, "Evidence-based genealogical research, for people who want to be able to show their work.")), cols.map(([t, ls]) => /*#__PURE__*/React.createElement("div", {
    key: t
  }, /*#__PURE__*/React.createElement("div", {
    className: "eyebrow"
  }, t), /*#__PURE__*/React.createElement("div", {
    style: {
      display: 'flex',
      flexDirection: 'column',
      gap: 'var(--space-4)',
      marginTop: 'var(--space-6)'
    }
  }, ls.map(l => /*#__PURE__*/React.createElement("a", {
    key: l,
    href: "#",
    style: {
      fontSize: 'var(--text-body-sm)',
      color: 'var(--text-secondary)',
      textDecoration: 'none'
    }
  }, l)))))));
}
function Site() {
  const [theme, setTheme] = React.useState('light');
  React.useEffect(() => {
    document.documentElement.setAttribute('data-theme', theme);
  }, [theme]);
  return /*#__PURE__*/React.createElement("div", {
    style: {
      background: 'var(--surface-page)',
      minHeight: '100vh'
    }
  }, /*#__PURE__*/React.createElement(Nav, {
    theme: theme,
    onTheme: () => setTheme(t => t === 'light' ? 'dark' : 'light')
  }), /*#__PURE__*/React.createElement(Hero, null), /*#__PURE__*/React.createElement(Grades, null), /*#__PURE__*/React.createElement(Features, null), /*#__PURE__*/React.createElement(Closing, null), /*#__PURE__*/React.createElement(Footer, null));
}
const siteRoot = document.getElementById('root');
if (siteRoot) ReactDOM.createRoot(siteRoot).render(/*#__PURE__*/React.createElement(Site, null));
})(); } catch (e) { __ds_ns.__errors.push({ path: "ui_kits/site/Site.jsx", error: String((e && e.message) || e) }); }

__ds_ns.Badge = __ds_scope.Badge;

__ds_ns.Button = __ds_scope.Button;

__ds_ns.Card = __ds_scope.Card;

__ds_ns.Icon = __ds_scope.Icon;

__ds_ns.IconButton = __ds_scope.IconButton;

__ds_ns.Tag = __ds_scope.Tag;

__ds_ns.Tooltip = __ds_scope.Tooltip;

__ds_ns.Callout = __ds_scope.Callout;

__ds_ns.Dialog = __ds_scope.Dialog;

__ds_ns.EmptyState = __ds_scope.EmptyState;

__ds_ns.Toast = __ds_scope.Toast;

__ds_ns.Checkbox = __ds_scope.Checkbox;

__ds_ns.Field = __ds_scope.Field;

__ds_ns.Input = __ds_scope.Input;

__ds_ns.Radio = __ds_scope.Radio;

__ds_ns.Select = __ds_scope.Select;

__ds_ns.Switch = __ds_scope.Switch;

__ds_ns.Breadcrumbs = __ds_scope.Breadcrumbs;

__ds_ns.SidebarNav = __ds_scope.SidebarNav;

__ds_ns.Tabs = __ds_scope.Tabs;

__ds_ns.EvidenceBadge = __ds_scope.EvidenceBadge;

__ds_ns.FactRow = __ds_scope.FactRow;

__ds_ns.PersonChip = __ds_scope.PersonChip;

__ds_ns.SourceCitation = __ds_scope.SourceCitation;

})();
