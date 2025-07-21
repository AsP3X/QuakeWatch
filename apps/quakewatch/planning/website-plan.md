# QuakeWatch - Website Plan

## Project Overview
The QuakeWatch Website is a modern Next.js application within the QuakeWatch monorepo, designed to display real-time earthquake data with interactive maps, user authentication, dashboards, and analytics. This application serves as the presentation layer for the entire QuakeWatch system, providing users with comprehensive access to earthquake information and seismic data.

## Application Architecture

### Frontend Architecture (Next.js 14+)

#### Core Framework
- **Framework**: Next.js 14+ with App Router
- **Language**: TypeScript for type safety
- **Styling**: Tailwind CSS with custom design system
- **State Management**: React Context API or Zustand
- **Build Tool**: Turbopack for fast development
- **Authentication**: NextAuth.js for user authentication
- **Monorepo**: Part of QuakeWatch monorepo with shared dependencies

#### Key Features
- **Server-Side Rendering (SSR)**: For SEO and performance
- **Static Site Generation (SSG)**: For static content
- **Incremental Static Regeneration (ISR)**: For dynamic data updates
- **API Routes**: Backend functionality within Next.js
- **Image Optimization**: Next.js Image component
- **Font Optimization**: Next.js Font optimization
- **Authentication**: Secure user authentication and authorization
- **Real-time Updates**: WebSocket integration for live earthquake data

### Component Architecture

#### 1. Layout Components
**Purpose**: Provide consistent structure across pages

**Components**:
- **Header**: Navigation, branding, user controls, authentication status
- **Footer**: Links, copyright, additional information
- **Sidebar**: Filters, navigation, quick actions (for dashboard pages)
- **Main Layout**: Page wrapper with responsive design
- **Auth Layout**: Protected route wrapper

**Implementation**:
```typescript
// Example layout structure
interface LayoutProps {
  children: React.ReactNode;
  title?: string;
  description?: string;
  requireAuth?: boolean;
  requireAdmin?: boolean;
}

const MainLayout: React.FC<LayoutProps> = ({ children, title, description, requireAuth, requireAdmin }) => {
  return (
    <div className="min-h-screen bg-bg-primary">
      <Header />
      <main className="container mx-auto px-4 py-8">
        {children}
      </main>
      <Footer />
    </div>
  );
};
```

#### 2. Map Components
**Purpose**: Display interactive earthquake data visualization

**Components**:
- **Map Container**: Main map wrapper with controls
- **Earthquake Layer**: Display earthquake points with magnitude-based circles
- **Time Slider**: Control for viewing earthquakes from different time periods (up to 12 hours back)
- **Legend**: Map legend with magnitude ranges and colors
- **Map Controls**: Zoom, pan, layer toggle controls
- **Popup Components**: Information popups for earthquake details
- **Magnitude Filter**: Filter earthquakes by magnitude range

**Map Libraries**:
- **Primary**: Leaflet.js with React-Leaflet
- **Alternative**: Mapbox GL JS
- **Styling**: Custom map styles with geological theme

**Implementation**:
```typescript
// Example map component
interface EarthquakeMapProps {
  earthquakes: Earthquake[];
  center: [number, number];
  zoom: number;
  timeRange: number; // hours back from now
  magnitudeFilter: [number, number];
  onEarthquakeClick: (earthquake: Earthquake) => void;
}

const EarthquakeMap: React.FC<EarthquakeMapProps> = ({ 
  earthquakes, 
  center, 
  zoom, 
  timeRange,
  magnitudeFilter,
  onEarthquakeClick 
}) => {
  return (
    <MapContainer center={center} zoom={zoom} className="h-full w-full">
      <TileLayer url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png" />
      <EarthquakeLayer 
        earthquakes={earthquakes} 
        onEarthquakeClick={onEarthquakeClick}
        magnitudeFilter={magnitudeFilter}
      />
      <TimeSlider timeRange={timeRange} />
      <MapLegend />
      <MapControls />
    </MapContainer>
  );
};
```

#### 3. Authentication and Authorization Components
**Purpose**: Handle user authentication and permission-based access control

**Components**:
- **Login Form**: User login interface
- **Register Form**: User registration interface
- **Password Reset**: Password reset functionality
- **Profile Management**: User profile editing
- **Auth Guards**: Route protection components
- **Permission Guards**: Permission-based access control
- **Role Management**: Role assignment and management
- **Permission Management**: Granular permission control

#### 4. Dashboard Components
**Purpose**: Provide user and admin dashboards

**User Dashboard Components**:
- **User Profile**: Personal information and preferences
- **Saved Earthquakes**: User's saved earthquake alerts
- **Notification Settings**: Alert preferences
- **Activity History**: Recent interactions

**Admin Dashboard Components**:
- **User Management**: View, edit, delete user accounts
- **Statistics Overview**: System usage statistics
- **Earthquake Analytics**: Data analysis and insights
- **System Health**: Application and database health
- **Data Management**: Earthquake data management tools

#### 5. Data Display Components
**Purpose**: Present earthquake data in various formats

**Components**:
- **Earthquake List**: Tabular display of earthquake data
- **Earthquake Card**: Individual earthquake information cards
- **Data Table**: Sortable, filterable data table
- **Statistics Cards**: Key metrics and statistics
- **Charts and Graphs**: Data visualization components
- **Real-time Feed**: Live earthquake updates

**Data Visualization Libraries**:
- **Charts**: Chart.js with react-chartjs-2
- **Advanced Charts**: D3.js for custom visualizations
- **Tables**: React Table for data tables

#### 6. Interactive Components
**Purpose**: User interaction and data manipulation

**Components**:
- **Search Bar**: Global search functionality
- **Filters**: Advanced filtering options (magnitude, location, time)
- **Sort Controls**: Data sorting controls
- **Export Tools**: Data export functionality
- **Settings Panel**: User preferences and settings
- **Notification System**: Real-time alerts and notifications

### Backend API (Next.js API Routes)

#### API Structure
```
/api/
├── auth/
│   ├── login.ts           # User login
│   ├── register.ts        # User registration
│   ├── logout.ts          # User logout
│   ├── profile.ts         # User profile management
│   └── reset-password.ts  # Password reset
├── permissions/
│   ├── check.ts           # Check user permissions
│   ├── list.ts            # List all permissions
│   ├── grant.ts           # Grant permission to user
│   └── revoke.ts          # Revoke permission from user
├── roles/
│   ├── list.ts            # List all roles
│   ├── [id].ts            # Role management
│   ├── assign.ts          # Assign role to user
│   └── permissions.ts     # Role permissions management
├── earthquakes/
│   ├── [id].ts            # Individual earthquake data
│   ├── search.ts          # Search functionality
│   ├── filter.ts          # Filtered results
│   ├── recent.ts          # Recent earthquakes (last 12 hours)
│   ├── export.ts          # Data export
│   └── real-time.ts       # Real-time earthquake feed
├── users/
│   ├── [id].ts            # User management (admin)
│   ├── list.ts            # User list (admin)
│   ├── statistics.ts      # User statistics (admin)
│   └── permissions.ts     # User permissions management
├── admin/
│   ├── dashboard.ts       # Admin dashboard data
│   ├── statistics.ts      # System statistics
│   ├── system-health.ts   # System health check
│   └── audit-logs.ts      # Permission audit logs
├── statistics/
│   ├── overview.ts        # General statistics
│   ├── regions.ts         # Regional statistics
│   └── trends.ts          # Trend analysis
└── system/
    ├── health.ts          # System health check
    └── status.ts          # Data freshness status
```

#### Database Integration
- **ORM**: Prisma for type-safe database queries
- **Connection Pooling**: Efficient database connections
- **Caching**: Redis for API response caching
- **Query Optimization**: Optimized database queries
- **Authentication**: NextAuth.js with database sessions
- **Authorization**: Role-based access control (RBAC) with granular permissions

## Page Structure

### 1. Landing Page (`/`) - Interactive Earthquake Map
**Purpose**: Main landing page with interactive earthquake map

**Features**:
- Full-screen interactive earthquake map
- Real-time earthquake data display
- Time slider (up to 12 hours back)
- Magnitude-based circle visualization
- Search and filter controls
- Legend and information panel
- Quick access to user dashboard and admin panel
- Responsive design for all devices

**Layout**:
```typescript
const LandingPage = () => {
  return (
    <MainLayout title="QuakeWatch - Real-time Earthquake Monitoring">
      <div className="h-screen">
        <EarthquakeMap 
          earthquakes={earthquakes}
          center={[0, 0]}
          zoom={2}
          timeRange={12}
          magnitudeFilter={[0, 10]}
          onEarthquakeClick={handleEarthquakeClick}
        />
        <MapControls />
        <QuickAccessPanel />
      </div>
    </MainLayout>
  );
};
```

### 2. Authentication Pages

#### Login Page (`/auth/login`)
**Purpose**: User authentication

**Features**:
- Email/password login
- Social login options (Google, GitHub)
- Remember me functionality
- Password reset link
- Registration link

#### Register Page (`/auth/register`)
**Purpose**: New user registration

**Features**:
- Email/password registration
- Email verification
- Terms and conditions
- Privacy policy acceptance

#### Password Reset Page (`/auth/reset-password`)
**Purpose**: Password recovery

**Features**:
- Email-based password reset
- Secure token validation
- New password setup

### 3. User Dashboard (`/dashboard`)
**Purpose**: User's personal dashboard (protected route)

**Features**:
- User profile management
- Saved earthquake alerts
- Notification preferences
- Activity history
- Quick access to map
- Personal statistics

### 4. Admin Dashboard (`/admin`)
**Purpose**: Administrative interface (admin-only route)

**Features**:
- User account management
- Role and permission management
- System statistics and analytics
- Earthquake data management
- System health monitoring
- Data export tools
- User activity logs
- Permission audit logs
- Role assignment interface

### 5. Earthquake Detail Page (`/earthquakes/[id]`)
**Purpose**: Detailed information about specific earthquakes

**Features**:
- Comprehensive earthquake information
- Historical data
- Related earthquakes
- Map visualization
- Technical specifications
- User comments and ratings

### 6. Analytics Page (`/analytics`)
**Purpose**: Data analysis and insights

**Features**:
- Statistical overview
- Trend analysis
- Regional comparisons
- Interactive charts
- Custom date ranges
- Export capabilities

### 7. Role Management Page (`/admin/roles`)
**Purpose**: Role and permission management interface (admin-only)

**Features**:
- Create, edit, and delete roles
- Assign permissions to roles
- View role assignments
- Permission matrix view
- Role hierarchy management
- Bulk permission operations

### 8. User Permissions Page (`/admin/users/[id]/permissions`)
**Purpose**: Individual user permission management (admin-only)

**Features**:
- View user's current permissions
- Grant or revoke individual permissions
- Set permission expiration dates
- View permission history
- Bulk permission operations
- Permission override management

### 9. About Page (`/about`)
**Purpose**: Information about the project and data sources

**Features**:
- Project description
- Data source information
- Methodology
- Team information
- Contact details

## Database Schema Updates

### User Management Tables

#### Users Table
```sql
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    role VARCHAR(20) DEFAULT 'user' CHECK (role IN ('user', 'admin')),
    email_verified BOOLEAN DEFAULT FALSE,
    email_verification_token VARCHAR(255),
    reset_password_token VARCHAR(255),
    reset_password_expires TIMESTAMP,
    preferences JSONB DEFAULT '{}',
    last_login TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_created_at ON users(created_at);
```

#### User Sessions Table
```sql
CREATE TABLE user_sessions (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    session_token VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Indexes
CREATE INDEX idx_user_sessions_user_id ON user_sessions(user_id);
CREATE INDEX idx_user_sessions_token ON user_sessions(session_token);
CREATE INDEX idx_user_sessions_expires ON user_sessions(expires_at);
```

### Earthquake Data Tables

#### Earthquakes Table
```sql
CREATE TABLE earthquakes (
    id SERIAL PRIMARY KEY,
    external_id VARCHAR(100) UNIQUE NOT NULL,
    magnitude DECIMAL(3,1) NOT NULL,
    magnitude_type VARCHAR(10),
    depth DECIMAL(8,3),
    latitude DECIMAL(10,7) NOT NULL,
    longitude DECIMAL(10,7) NOT NULL,
    location VARCHAR(255),
    region VARCHAR(255),
    country VARCHAR(100),
    time TIMESTAMP NOT NULL,
    updated_time TIMESTAMP,
    
    -- Geographical data
    geometry GEOMETRY(POINT, 4326),
    
    -- Additional data
    place VARCHAR(500),
    type VARCHAR(50),
    status VARCHAR(50),
    tsunami BOOLEAN DEFAULT FALSE,
    felt_count INTEGER,
    significance INTEGER,
    
    -- Metadata
    source VARCHAR(100),
    raw_data JSONB,
    
    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Constraints
    CONSTRAINT valid_magnitude CHECK (magnitude >= 0),
    CONSTRAINT valid_depth CHECK (depth >= 0),
    CONSTRAINT valid_coordinates CHECK (latitude BETWEEN -90 AND 90 AND longitude BETWEEN -180 AND 180)
);

-- Indexes
CREATE INDEX idx_earthquakes_external_id ON earthquakes(external_id);
CREATE INDEX idx_earthquakes_magnitude ON earthquakes(magnitude);
CREATE INDEX idx_earthquakes_time ON earthquakes(time);
CREATE INDEX idx_earthquakes_geometry ON earthquakes USING GIST(geometry);
CREATE INDEX idx_earthquakes_location ON earthquakes(location);
CREATE INDEX idx_earthquakes_region ON earthquakes(region);
```

#### User Earthquakes Table (for saved earthquakes)
```sql
CREATE TABLE user_earthquakes (
    id SERIAL PRIMARY KEY,
    user_id INTEGER REFERENCES users(id) ON DELETE CASCADE,
    earthquake_id INTEGER REFERENCES earthquakes(id) ON DELETE CASCADE,
    saved_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    notes TEXT,
    
    UNIQUE(user_id, earthquake_id)
);

-- Indexes
CREATE INDEX idx_user_earthquakes_user_id ON user_earthquakes(user_id);
CREATE INDEX idx_user_earthquakes_earthquake_id ON user_earthquakes(earthquake_id);
```

## Implementation Phases

### Phase 1: Foundation Setup (Week 1)

#### Project Initialization
1. **Next.js Setup**
   - Initialize Next.js 14+ project with TypeScript
   - Configure Tailwind CSS
   - Set up ESLint, Prettier, and Husky
   - Configure environment variables

2. **Authentication Setup**
   - Install and configure NextAuth.js
   - Set up database authentication tables
   - Create authentication components
   - Implement protected routes

3. **Design System Implementation**
   - Implement color scheme from design document
   - Create base components (Button, Input, Card, etc.)
   - Set up typography system
   - Create responsive utilities

4. **Database Integration**
   - Set up Prisma ORM
   - Create database schema (including user and earthquake tables)
   - Implement connection pooling
   - Set up data models

### Phase 2: Core Components (Week 2-3)

#### Week 2: Layout and Authentication
1. **Layout Components**
   - Create main layout wrapper
   - Implement header with navigation and auth status
   - Build footer component
   - Create responsive sidebar for dashboards

2. **Authentication Pages**
   - Implement login page
   - Create registration page
   - Build password reset functionality
   - Add authentication guards

#### Week 3: Map and Data Display
1. **Map Components**
   - Integrate Leaflet.js
   - Create earthquake map component
   - Implement magnitude-based circle visualization
   - Add time slider functionality

2. **Data Components**
   - Create earthquake list component
   - Build earthquake detail cards
   - Implement data table
   - Add statistics cards

### Phase 3: Dashboards and User Features (Week 4-5)

#### Week 4: User Dashboard
1. **User Dashboard**
   - Create user profile management
   - Implement saved earthquakes functionality
   - Add notification preferences
   - Build activity history

2. **API Routes**
   - Create user management API endpoints
   - Implement earthquake data API
   - Add search and filtering capabilities
   - Set up data export functionality

#### Week 5: Admin Dashboard
1. **Admin Dashboard**
   - Create user management interface
   - Implement system statistics
   - Add earthquake data management
   - Build system health monitoring

2. **Advanced Features**
   - Add real-time earthquake updates
   - Implement notification system
   - Create data analytics tools
   - Add export and reporting features

### Phase 4: Advanced Features (Week 6)

#### Analytics and Visualization
1. **Charts and Graphs**
   - Integrate Chart.js
   - Create statistical visualizations
   - Build trend analysis charts
   - Implement interactive graphs

2. **Advanced Map Features**
   - Add clustering for large datasets
   - Implement custom map styles
   - Create advanced filtering
   - Add map export functionality

### Phase 5: Optimization and Polish (Week 7-8)

#### Week 7: Performance and SEO
1. **Performance Optimization**
   - Implement code splitting
   - Add image optimization
   - Optimize bundle size
   - Add caching strategies

2. **SEO and Accessibility**
   - Add meta tags and structured data
   - Implement accessibility features
   - Add keyboard navigation
   - Test with screen readers

#### Week 8: Testing and Deployment
1. **Testing**
   - Write unit tests
   - Add integration tests
   - Perform end-to-end testing
   - Conduct user testing

2. **Deployment**
   - Set up production environment
   - Configure CI/CD pipeline
   - Implement monitoring
   - Create deployment documentation

## Technical Stack

### Frontend Framework
- **Framework**: Next.js 14+
- **Language**: TypeScript
- **Styling**: Tailwind CSS
- **State Management**: React Context API / Zustand
- **Build Tool**: Turbopack
- **Authentication**: NextAuth.js

### Map and Visualization
- **Maps**: Leaflet.js with React-Leaflet
- **Charts**: Chart.js with react-chartjs-2
- **Advanced Charts**: D3.js
- **Tables**: React Table

### Backend and Database
- **API**: Next.js API Routes
- **Database**: PostgreSQL with PostGIS
- **ORM**: Prisma
- **Caching**: Redis
- **Authentication**: NextAuth.js with database sessions
- **Authorization**: Custom RBAC middleware with permission checking

### Development Tools
- **Code Quality**: ESLint, Prettier, Husky
- **Testing**: Jest, React Testing Library, Playwright
- **Type Checking**: TypeScript
- **Package Management**: npm or yarn

### Deployment and DevOps
- **Hosting**: Vercel, Netlify, or AWS
- **CI/CD**: GitHub Actions
- **Monitoring**: Vercel Analytics, Sentry
- **Performance**: Core Web Vitals monitoring

## Monorepo Integration

### Application Structure
The web application is organized as part of the QuakeWatch monorepo:
- Located in `apps/web/` directory
- Shares common dependencies with the scraper application
- Uses shared packages for database access and utilities
- Maintains independent configuration and deployment

### Shared Dependencies
- Database connection and Prisma client
- Common data types and interfaces
- Shared utility functions
- Configuration management
- Logging and monitoring utilities

### Development Workflow
- Can be developed and tested independently
- Shares the same database with the scraper application
- Uses workspace dependencies for local development
- Can be deployed separately or together with other applications

## Configuration
```bash
# Database Configuration
DATABASE_URL=postgresql://user:password@localhost:5432/quakewatch
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_DB=quakewatch
POSTGRES_USER=quakewatch_user
POSTGRES_PASSWORD=secure_password

# Authentication Configuration
NEXTAUTH_SECRET=your-nextauth-secret
NEXTAUTH_URL=http://localhost:3000
GOOGLE_CLIENT_ID=your-google-client-id
GOOGLE_CLIENT_SECRET=your-google-client-secret
GITHUB_CLIENT_ID=your-github-client-id
GITHUB_CLIENT_SECRET=your-github-client-secret

# Email Configuration (for password reset)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your-email@gmail.com
SMTP_PASSWORD=your-app-password

# API Configuration
NEXT_PUBLIC_API_BASE_URL=http://localhost:3000/api
NEXT_PUBLIC_MAP_TILE_URL=https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png

# Application Configuration
NEXT_PUBLIC_APP_NAME=QuakeWatch
NEXT_PUBLIC_APP_DESCRIPTION=Real-time Earthquake Monitoring System
NEXT_PUBLIC_APP_URL=http://localhost:3000

# External Services
REDIS_URL=redis://localhost:6379
SENTRY_DSN=your-sentry-dsn
GOOGLE_ANALYTICS_ID=your-ga-id

# Earthquake Data API
EARTHQUAKE_API_URL=https://earthquake.usgs.gov/earthquakes/feed/v1.0
EARTHQUAKE_API_KEY=your-api-key-if-required
```

### Configuration Files
- `next.config.js` - Next.js configuration
- `tailwind.config.js` - Tailwind CSS configuration
- `prisma/schema.prisma` - Database schema
- `tsconfig.json` - TypeScript configuration
- `jest.config.js` - Testing configuration
- `nextauth.config.js` - NextAuth.js configuration
- `middleware.ts` - Permission middleware configuration

## Permission System Implementation

### Permission Middleware
```typescript
// middleware.ts - Next.js middleware for permission checking
import { NextRequest, NextResponse } from 'next/server';
import { getToken } from 'next-auth/jwt';

export async function middleware(request: NextRequest) {
  const token = await getToken({ req: request });
  
  // Public routes that don't require authentication
  const publicRoutes = ['/', '/auth/login', '/auth/register', '/about'];
  const isPublicRoute = publicRoutes.some(route => request.nextUrl.pathname.startsWith(route));
  
  if (isPublicRoute) {
    return NextResponse.next();
  }
  
  // Check authentication
  if (!token) {
    return NextResponse.redirect(new URL('/auth/login', request.url));
  }
  
  // Check permissions for protected routes
  const userPermissions = await getUserPermissions(token.sub as string);
  
  // Admin routes require admin permissions
  if (request.nextUrl.pathname.startsWith('/admin')) {
    if (!userPermissions.includes('admin:dashboard')) {
      return NextResponse.redirect(new URL('/dashboard', request.url));
    }
  }
  
  // User management routes require specific permissions
  if (request.nextUrl.pathname.startsWith('/api/users')) {
    if (!userPermissions.includes('users:read')) {
      return NextResponse.json({ error: 'Insufficient permissions' }, { status: 403 });
    }
  }
  
  return NextResponse.next();
}

export const config = {
  matcher: [
    '/admin/:path*',
    '/dashboard/:path*',
    '/api/users/:path*',
    '/api/admin/:path*',
    '/api/permissions/:path*',
    '/api/roles/:path*'
  ]
};
```

### Permission Hooks
```typescript
// hooks/usePermissions.ts - React hook for permission checking
import { useSession } from 'next-auth/react';
import { useState, useEffect } from 'react';

export function usePermissions() {
  const { data: session } = useSession();
  const [permissions, setPermissions] = useState<string[]>([]);
  const [loading, setLoading] = useState(true);
  
  useEffect(() => {
    if (session?.user?.id) {
      fetchUserPermissions(session.user.id)
        .then(setPermissions)
        .finally(() => setLoading(false));
    } else {
      setPermissions([]);
      setLoading(false);
    }
  }, [session]);
  
  const hasPermission = (permission: string) => {
    return permissions.includes(permission);
  };
  
  const hasAnyPermission = (permissionList: string[]) => {
    return permissionList.some(permission => permissions.includes(permission));
  };
  
  const hasAllPermissions = (permissionList: string[]) => {
    return permissionList.every(permission => permissions.includes(permission));
  };
  
  return {
    permissions,
    loading,
    hasPermission,
    hasAnyPermission,
    hasAllPermissions
  };
}
```

### Permission Components
```typescript
// components/PermissionGuard.tsx - Component for permission-based rendering
import { usePermissions } from '../hooks/usePermissions';

interface PermissionGuardProps {
  permission: string;
  fallback?: React.ReactNode;
  children: React.ReactNode;
}

export function PermissionGuard({ permission, fallback, children }: PermissionGuardProps) {
  const { hasPermission, loading } = usePermissions();
  
  if (loading) {
    return <div>Loading...</div>;
  }
  
  if (!hasPermission(permission)) {
    return fallback || null;
  }
  
  return <>{children}</>;
}

// components/RoleGuard.tsx - Component for role-based rendering
interface RoleGuardProps {
  roles: string[];
  fallback?: React.ReactNode;
  children: React.ReactNode;
}

export function RoleGuard({ roles, fallback, children }: RoleGuardProps) {
  const { session } = useSession();
  const userRole = session?.user?.role;
  
  if (!userRole || !roles.includes(userRole)) {
    return fallback || null;
  }
  
  return <>{children}</>;
}
```

## Performance Optimization

### Frontend Optimization
- **Code Splitting**: Dynamic imports for large components
- **Image Optimization**: Next.js Image component
- **Font Optimization**: Next.js Font optimization
- **Bundle Analysis**: Regular bundle size monitoring
- **Lazy Loading**: Implement lazy loading for components

### Backend Optimization
- **Database Queries**: Optimize database queries
- **Caching**: Implement Redis caching
- **API Response**: Compress API responses
- **Connection Pooling**: Efficient database connections

### Map Performance
- **Tile Loading**: Optimize map tile loading
- **Clustering**: Implement point clustering for large datasets
- **Viewport Loading**: Load data based on map viewport
- **Memory Management**: Proper cleanup of map resources

## Accessibility and SEO

### Accessibility Features
- **Keyboard Navigation**: Full keyboard accessibility
- **Screen Reader Support**: ARIA labels and semantic HTML
- **Color Contrast**: WCAG 2.1 AA compliance
- **Focus Management**: Proper focus indicators
- **Alternative Text**: Descriptive alt text for images

### SEO Optimization
- **Meta Tags**: Comprehensive meta tag implementation
- **Structured Data**: JSON-LD structured data
- **Sitemap**: Automatic sitemap generation
- **Robots.txt**: Proper robots.txt configuration
- **Open Graph**: Social media sharing optimization

## Testing Strategy

### Unit Testing
- **Component Testing**: Test individual components
- **Utility Testing**: Test utility functions
- **API Testing**: Test API route handlers
- **Type Testing**: TypeScript type checking

### Integration Testing
- **Page Testing**: Test complete page functionality
- **API Integration**: Test API integration
- **Database Testing**: Test database operations
- **Map Testing**: Test map functionality
- **Authentication Testing**: Test auth flows

### End-to-End Testing
- **User Flows**: Test complete user journeys
- **Cross-Browser Testing**: Test across different browsers
- **Mobile Testing**: Test on mobile devices
- **Performance Testing**: Test performance metrics

## Deployment Strategy

### Development Environment
- **Local Development**: Docker Compose setup
- **Hot Reloading**: Fast development experience
- **Environment Variables**: Local environment configuration
- **Database**: Local PostgreSQL instance

### Staging Environment
- **Preview Deployments**: Automatic staging deployments
- **Testing**: Comprehensive testing in staging
- **Data**: Staging database with sample data
- **Monitoring**: Staging environment monitoring

### Production Environment
- **Hosting**: Vercel or similar platform
- **Database**: Managed PostgreSQL service
- **CDN**: Global content delivery network
- **Monitoring**: Production monitoring and alerting

## Success Metrics

### Performance Metrics
- **Page Load Time**: < 3 seconds for initial load
- **Time to Interactive**: < 5 seconds
- **Core Web Vitals**: Good scores across all metrics
- **Bundle Size**: < 500KB initial bundle

### User Experience Metrics
- **User Engagement**: Time spent on site
- **Map Interaction**: Map usage statistics
- **Search Usage**: Search functionality usage
- **Export Usage**: Data export functionality usage
- **User Registration**: New user signup rate
- **User Retention**: Returning user rate

### Technical Metrics
- **Uptime**: > 99.9% availability
- **Error Rate**: < 0.1% error rate
- **API Response Time**: < 500ms average
- **Database Performance**: < 100ms query time
- **Authentication Success Rate**: > 99% successful logins

## Future Enhancements

### Phase 2 Features
- **Real-time Notifications**: Push notifications for earthquakes
- **Mobile App**: React Native mobile application
- **Advanced Analytics**: Machine learning insights
- **API Documentation**: Comprehensive API documentation
- **Multi-language Support**: Internationalization

### Phase 3 Features
- **Advanced Visualizations**: 3D earthquake models
- **Data Integration**: Multiple earthquake data sources
- **Predictive Analytics**: Earthquake prediction models
- **Community Features**: User contributions and discussions
- **Emergency Alerts**: Integration with emergency services

## Conclusion

The QuakeWatch Website provides a modern, accessible, and performant interface for exploring real-time earthquake data. The Next.js framework ensures excellent developer experience and optimal performance, while the comprehensive component architecture allows for easy maintenance and future enhancements.

The website serves as the primary interface for users to interact with earthquake data, providing both simple access for casual users and advanced features for researchers and professionals in the field of seismology. The addition of user authentication and dashboards creates a personalized experience while the interactive map serves as an engaging landing page that immediately showcases the system's capabilities. 