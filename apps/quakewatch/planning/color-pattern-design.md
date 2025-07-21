# QuakeWatch - Color Pattern Design

## Design Philosophy
The color scheme for QuakeWatch is inspired by geological formations, seismic activity, and earth sciences while maintaining excellent accessibility and user experience. The palette emphasizes trust, stability, and scientific accuracy.

## Primary Color Palette

### Core Colors
```css
/* Primary Brand Colors */
--primary-earth: #8B4513;        /* Rich brown - represents earth/soil */
--primary-stone: #696969;        /* Medium gray - represents rock formations */
--primary-sand: #F4A460;         /* Sandy beige - represents geological layers */

/* Secondary Accent Colors */
--accent-rust: #CD5C5C;          /* Rust red - represents iron oxide in rocks */
--accent-copper: #B87333;        /* Copper brown - represents mineral deposits */
--accent-slate: #708090;         /* Slate gray - represents sedimentary rock */
```

### Seismic Activity Colors
```css
/* Fault and Seismic Colors */
--fault-active: #FF4500;         /* Bright orange - active fault lines */
--fault-inactive: #A0522D;       /* Sienna brown - inactive fault lines */
--earthquake-high: #DC143C;      /* Crimson red - high magnitude events */
--earthquake-medium: #FF8C00;    /* Dark orange - medium magnitude events */
--earthquake-low: #FFD700;       /* Gold yellow - low magnitude events */
```

## Extended Color Palette

### Neutral Colors
```css
/* Background and Text Colors */
--bg-primary: #FAFAFA;           /* Off-white - main background */
--bg-secondary: #F5F5F5;         /* Light gray - secondary background */
--bg-tertiary: #E8E8E8;          /* Medium gray - tertiary background */
--bg-dark: #2F2F2F;              /* Dark gray - dark mode background */

--text-primary: #1A1A1A;         /* Near black - primary text */
--text-secondary: #4A4A4A;       /* Dark gray - secondary text */
--text-muted: #6B6B6B;           /* Medium gray - muted text */
--text-light: #FFFFFF;           /* White - light text on dark backgrounds */
```

### Status and Alert Colors
```css
/* Status Indicators */
--success: #228B22;              /* Forest green - success states */
--warning: #FFA500;              /* Orange - warning states */
--error: #DC143C;                /* Crimson red - error states */
--info: #4169E1;                 /* Royal blue - informational states */

/* Alert Levels */
--alert-critical: #8B0000;       /* Dark red - critical alerts */
--alert-high: #FF0000;           /* Red - high priority */
--alert-medium: #FF8C00;         /* Orange - medium priority */
--alert-low: #FFD700;            /* Gold - low priority */
```

## Color Usage Guidelines

### Primary Application
- **Header/Navigation**: Use `--primary-earth` for main navigation background
- **Primary Buttons**: Use `--primary-earth` with `--text-light` text
- **Secondary Buttons**: Use `--primary-stone` with `--text-light` text
- **Links**: Use `--accent-rust` for hover states, `--primary-earth` for default

### Map Visualization
- **Active Faults**: `--fault-active` for currently active fault lines
- **Inactive Faults**: `--fault-inactive` for historical fault lines
- **Earthquake Magnitude**:
  - 0-2.9: `--earthquake-low`
  - 3.0-4.9: `--earthquake-medium`
  - 5.0+: `--earthquake-high`

### Data Visualization
- **Charts and Graphs**: Use the seismic activity colors for magnitude-based visualizations
- **Progress Bars**: Use `--primary-earth` to `--accent-rust` gradients
- **Status Indicators**: Use the status colors for system health and data freshness

## Accessibility Considerations

### Contrast Ratios
All color combinations meet WCAG 2.1 AA standards:
- **Normal Text**: Minimum 4.5:1 contrast ratio
- **Large Text**: Minimum 3:1 contrast ratio
- **UI Components**: Minimum 3:1 contrast ratio

### Color Blindness Support
- Avoid relying solely on color to convey information
- Use patterns, icons, and text labels in addition to color
- Test with color blindness simulators
- Provide alternative indicators for all color-coded information

### High Contrast Mode
```css
/* High contrast mode overrides */
@media (prefers-contrast: high) {
  --primary-earth: #5D2906;
  --text-primary: #000000;
  --bg-primary: #FFFFFF;
}
```

## Dark Mode Support

### Dark Theme Colors
```css
/* Dark Mode Palette */
--dark-bg-primary: #1A1A1A;      /* Dark background */
--dark-bg-secondary: #2D2D2D;    /* Secondary dark background */
--dark-bg-tertiary: #404040;     /* Tertiary dark background */

--dark-text-primary: #FFFFFF;    /* White text */
--dark-text-secondary: #E0E0E0;  /* Light gray text */
--dark-text-muted: #B0B0B0;      /* Muted text */

--dark-primary-earth: #A0522D;   /* Lighter earth tone for dark mode */
--dark-primary-stone: #808080;   /* Lighter stone tone for dark mode */
```

## Implementation Guidelines

### CSS Custom Properties
```css
:root {
  /* Primary Colors */
  --primary-earth: #8B4513;
  --primary-stone: #696969;
  --primary-sand: #F4A460;
  
  /* Secondary Colors */
  --accent-rust: #CD5C5C;
  --accent-copper: #B87333;
  --accent-slate: #708090;
  
  /* Seismic Colors */
  --fault-active: #FF4500;
  --fault-inactive: #A0522D;
  --earthquake-high: #DC143C;
  --earthquake-medium: #FF8C00;
  --earthquake-low: #FFD700;
  
  /* Neutral Colors */
  --bg-primary: #FAFAFA;
  --bg-secondary: #F5F5F5;
  --bg-tertiary: #E8E8E8;
  --text-primary: #1A1A1A;
  --text-secondary: #4A4A4A;
  --text-muted: #6B6B6B;
  --text-light: #FFFFFF;
  
  /* Status Colors */
  --success: #228B22;
  --warning: #FFA500;
  --error: #DC143C;
  --info: #4169E1;
}

/* Dark Mode */
@media (prefers-color-scheme: dark) {
  :root {
    --bg-primary: #1A1A1A;
    --bg-secondary: #2D2D2D;
    --bg-tertiary: #404040;
    --text-primary: #FFFFFF;
    --text-secondary: #E0E0E0;
    --text-muted: #B0B0B0;
    --primary-earth: #A0522D;
    --primary-stone: #808080;
  }
}
```

### Component Examples

#### Button Styles
```css
.btn-primary {
  background-color: var(--primary-earth);
  color: var(--text-light);
  border: none;
  padding: 12px 24px;
  border-radius: 6px;
  transition: background-color 0.2s ease;
}

.btn-primary:hover {
  background-color: var(--accent-rust);
}

.btn-secondary {
  background-color: var(--primary-stone);
  color: var(--text-light);
  border: 1px solid var(--primary-earth);
}
```

#### Map Legend
```css
.legend-item {
  display: flex;
  align-items: center;
  margin: 8px 0;
}

.legend-color {
  width: 20px;
  height: 20px;
  margin-right: 12px;
  border-radius: 3px;
}

.legend-fault-active { background-color: var(--fault-active); }
.legend-fault-inactive { background-color: var(--fault-inactive); }
.legend-earthquake-high { background-color: var(--earthquake-high); }
.legend-earthquake-medium { background-color: var(--earthquake-medium); }
.legend-earthquake-low { background-color: var(--earthquake-low); }
```

## Brand Identity

### Logo Colors
- **Primary Logo**: Use `--primary-earth` as the main color
- **Secondary Logo**: Use `--primary-stone` for monochrome versions
- **Accent Elements**: Use `--accent-rust` for highlights

### Typography Colors
- **Headings**: `--text-primary` for main headings
- **Body Text**: `--text-secondary` for readable content
- **Captions**: `--text-muted` for supplementary information
- **Links**: `--accent-rust` with underline on hover

## Responsive Considerations

### Mobile Optimization
- Ensure touch targets have sufficient contrast
- Use larger color blocks for mobile interfaces
- Maintain readability on smaller screens
- Test color visibility in various lighting conditions

### Print-Friendly
```css
@media print {
  :root {
    --primary-earth: #000000;
    --text-primary: #000000;
    --bg-primary: #FFFFFF;
    --fault-active: #000000;
    --fault-inactive: #666666;
  }
}
```

## Testing Checklist

### Accessibility Testing
- [ ] Test with color blindness simulators
- [ ] Verify contrast ratios meet WCAG standards
- [ ] Test with screen readers
- [ ] Validate in high contrast mode
- [ ] Test with reduced motion preferences

### Cross-Platform Testing
- [ ] Test on different browsers (Chrome, Firefox, Safari, Edge)
- [ ] Test on mobile devices (iOS, Android)
- [ ] Test in different lighting conditions
- [ ] Verify dark mode functionality
- [ ] Test print layout

### User Experience Testing
- [ ] Conduct user testing with target audience
- [ ] Gather feedback on color preferences
- [ ] Test emotional response to color choices
- [ ] Validate color associations with geological concepts

## Conclusion

This color pattern design provides a comprehensive foundation for the QuakeWatch website that balances scientific accuracy with excellent user experience. The geological theme creates a professional and trustworthy appearance while the accessibility considerations ensure the platform is usable by all users.

The color system is designed to be flexible and scalable, allowing for future enhancements while maintaining consistency across all components of the application. 