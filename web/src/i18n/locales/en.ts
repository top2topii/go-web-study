import type ko from './ko'

// `typeof ko`로 타입을 지정하면 키가 빠지거나 오타가 나면 컴파일 에러가 난다.
const en: typeof ko = {
  app: {
    name: 'ShopAdmin',
    tagline: 'E-commerce Admin',
  },
  nav: {
    groups: {
      operations: 'Operations',
      management: 'Management',
      design: 'Design System',
    },
    dashboard: 'Dashboard',
    orders: 'Orders',
    products: 'Products',
    categories: 'Categories',
    customers: 'Customers',
    users: 'Admin Users',
    designOverview: 'Overview',
    home: 'Home',
    toggleSidebar: 'Toggle sidebar',
  },
  theme: {
    label: 'Theme',
    mode: 'Mode',
    light: 'Light',
    dark: 'Dark',
    system: 'System',
    brand: 'Brand color',
    brands: {
      blue: 'Blue',
      violet: 'Violet',
      emerald: 'Emerald',
      rose: 'Rose',
      orange: 'Orange',
      neutral: 'Neutral',
    },
  },
  lang: {
    label: 'Language',
  },
  user: {
    account: 'My account',
    logout: 'Log out',
    roles: {
      admin: 'Admin',
      viewer: 'Viewer',
    },
  },
  login: {
    title: 'Sign in',
    subtitle: 'Sign in with your admin account.',
    email: 'Email',
    password: 'Password',
    submit: 'Sign in',
    submitting: 'Signing in…',
    invalid: 'Invalid email or password.',
    demoHint: 'Demo accounts: admin@example.com / viewer@example.com (password: see make seed output)',
  },
  errors: {
    generic: 'Something went wrong. Please try again.',
    network: 'Cannot reach the server.',
    badRequest: 'Please check your input.',
    forbidden: 'You do not have permission to do this.',
    retry: 'Retry',
  },
  common: {
    comingSoon: 'Coming soon',
    comingSoonDesc: 'Planned for phase {{phase}}.',
  },
  dashboard: {
    title: 'Dashboard',
    subtitle: 'Last {{days}} days (excluding cancelled/refunded)',
    revenue: 'Revenue',
    orders: 'Orders',
    aov: 'Avg. order value',
    today: "Today's revenue",
    perDay: '{{value}} / day',
  },
  design: {
    title: 'Design System',
    subtitle: 'Component catalog built on shadcn/ui (Radix) + Tailwind CSS',
  },
  notFound: {
    title: 'Page not found',
    desc: 'Please check the address.',
    back: 'Back to dashboard',
  },
}

export default en
