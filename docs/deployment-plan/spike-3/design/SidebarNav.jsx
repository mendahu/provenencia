// Proposed replacement for the Provenencia design system's navigation/SidebarNav.jsx.
// Adds a `collapsed` (icon-only rail) mode so both sidebar states come from one component.
// Expanded: group label + icon + label + optional count.
// Collapsed: icons only, group labels dropped, groups separated by a hairline divider,
// destination name exposed via Tooltip + aria-label. States: idle / hover / selected / disabled.

const DS = () => window.ProvenenciaDesignSystem_0f6c1f || {};

function NavButton({ item, active, collapsed, onChange }) {
  const { Icon } = DS();
  const [hover, setHover] = React.useState(false);
  const [tip, setTip] = React.useState(null);
  const disabled = !!item.disabled;
  const bg = active
    ? 'var(--surface-selected)'
    : hover && !disabled
    ? 'var(--surface-hover)'
    : 'transparent';

  const button = (
    <button
      onClick={() => !disabled && onChange && onChange(item.id)}
      onMouseEnter={(e) => {
        setHover(true);
        if (collapsed) {
          const r = e.currentTarget.getBoundingClientRect();
          setTip({ left: r.right + 8, top: r.top + r.height / 2 });
        }
      }}
      onMouseLeave={() => { setHover(false); setTip(null); }}
      disabled={disabled}
      aria-current={active ? 'page' : undefined}
      aria-label={collapsed ? item.label : undefined}
      style={{
        display: 'flex',
        alignItems: 'center',
        justifyContent: collapsed ? 'center' : 'flex-start',
        gap: 'var(--space-5)',
        width: collapsed ? 32 : '100%',
        height: collapsed ? 32 : undefined,
        padding: collapsed ? 0 : '7px 10px',
        border: 0,
        cursor: disabled ? 'not-allowed' : 'pointer',
        textAlign: 'left',
        borderRadius: 'var(--radius-sm)',
        background: bg,
        color: active ? 'var(--text-primary)' : 'var(--text-secondary)',
        opacity: disabled ? 0.4 : 1,
        fontFamily: 'var(--font-body)',
        fontSize: 'var(--text-body-sm)',
        fontWeight: active ? 'var(--weight-semibold)' : 'var(--weight-regular)',
        transition: 'background var(--dur-fast) var(--ease-standard), color var(--dur-fast) var(--ease-standard)',
      }}
    >
      {item.icon && Icon ? (
        <Icon name={item.icon} size={collapsed ? 16 : 15} strokeWeight={active ? 'regular' : 'light'} />
      ) : null}
      {!collapsed && (
        <span style={{ flex: 1, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
          {item.label}
        </span>
      )}
      {!collapsed && item.count != null && (
        <span style={{ fontFamily: 'var(--font-mono)', fontSize: 'var(--text-micro)', color: 'var(--text-faint)' }}>
          {item.count}
        </span>
      )}
    </button>
  );

  if (!collapsed) return button;
  // position: fixed so the label is not clipped by the sidebar's own scroll container
  return (
    <span style={{ display: 'inline-flex', position: 'relative' }}>
      {button}
      {tip && (
        <span
          role="tooltip"
          style={{
            position: 'fixed', left: tip.left, top: tip.top, transform: 'translateY(-50%)',
            zIndex: 200, pointerEvents: 'none', whiteSpace: 'nowrap',
            background: 'var(--paper-900)', color: 'var(--paper-50)',
            fontFamily: 'var(--font-body)', fontSize: 'var(--text-micro)',
            letterSpacing: 'var(--tracking-wide)', padding: '4px 8px',
            borderRadius: 'var(--radius-xs)', boxShadow: 'var(--shadow-md)',
          }}
        >
          {item.label}
        </span>
      )}
    </span>
  );
}

function SidebarNav({ groups = [], value, onChange, collapsed = false, style, ...rest }) {
  return (
    <nav
      style={{
        display: 'flex',
        flexDirection: 'column',
        alignItems: collapsed ? 'center' : 'stretch',
        gap: collapsed ? 'var(--space-4)' : 'var(--space-8)',
        padding: collapsed ? '14px 0' : '16px 12px',
        ...style,
      }}
      {...rest}
    >
      {groups.map((g, gi) => (
        <React.Fragment key={gi}>
          {collapsed && gi > 0 && (
            <div style={{ width: 24, height: 1, background: 'var(--border-subtle)', margin: 'var(--space-1) 0' }} />
          )}
          <div
            style={{
              display: 'flex',
              flexDirection: 'column',
              alignItems: collapsed ? 'center' : 'stretch',
              gap: collapsed ? 'var(--space-2)' : 'var(--space-1)',
            }}
          >
            {g.label && !collapsed && (
              <div
                style={{
                  fontSize: 'var(--text-micro)',
                  letterSpacing: 'var(--tracking-caps)',
                  textTransform: 'uppercase',
                  color: 'var(--text-faint)',
                  fontWeight: 'var(--weight-semibold)',
                  padding: '0 10px var(--space-4)',
                }}
              >
                {g.label}
              </div>
            )}
            {g.items.map((it) => (
              <NavButton
                key={it.id}
                item={it}
                active={it.id === value}
                collapsed={collapsed}
                onChange={onChange}
              />
            ))}
          </div>
        </React.Fragment>
      ))}
    </nav>
  );
}

window.SidebarNav = SidebarNav;
module.exports = { SidebarNav };
