// 한국어가 기준(source) 언어다. en.ts는 이 객체와 같은 모양이어야 컴파일된다.
const ko = {
  app: {
    name: 'ShopAdmin',
    tagline: '이커머스 관리자',
  },
  nav: {
    groups: {
      operations: '운영',
      management: '관리',
      design: '디자인 시스템',
    },
    dashboard: '대시보드',
    orders: '주문',
    products: '상품',
    categories: '카테고리',
    customers: '고객',
    users: '관리자 계정',
    designOverview: '개요',
    home: '홈',
    toggleSidebar: '사이드바 열기/닫기',
  },
  theme: {
    label: '테마',
    mode: '모드',
    light: '라이트',
    dark: '다크',
    system: '시스템',
    brand: '브랜드 색',
    brands: {
      blue: '블루',
      violet: '바이올렛',
      emerald: '에메랄드',
      rose: '로즈',
      orange: '오렌지',
      neutral: '뉴트럴',
    },
  },
  lang: {
    label: '언어',
  },
  user: {
    account: '내 계정',
    logout: '로그아웃',
    roles: {
      admin: '관리자',
      viewer: '뷰어',
    },
  },
  login: {
    title: '로그인',
    subtitle: '관리자 계정으로 로그인하세요.',
    email: '이메일',
    password: '비밀번호',
    submit: '로그인',
    submitting: '로그인 중…',
    invalid: '이메일 또는 비밀번호가 올바르지 않습니다.',
    demoHint: '데모 계정: admin@example.com / viewer@example.com (비밀번호는 make seed 출력 참고)',
  },
  errors: {
    generic: '문제가 발생했습니다. 잠시 후 다시 시도해 주세요.',
    network: '서버에 연결할 수 없습니다.',
    badRequest: '입력값을 확인해 주세요.',
    forbidden: '이 작업을 수행할 권한이 없습니다.',
    retry: '다시 시도',
  },
  common: {
    comingSoon: '준비 중',
    comingSoonDesc: '{{phase}}단계에서 구현할 예정입니다.',
  },
  dashboard: {
    title: '대시보드',
    subtitle: '최근 {{days}}일 운영 현황 (취소·환불 제외)',
    revenue: '매출',
    orders: '주문 수',
    aov: '평균 주문 금액',
    today: '오늘 매출',
    perDay: '일평균 {{value}}',
  },
  design: {
    title: '디자인 시스템',
    subtitle: 'shadcn/ui(Radix) + Tailwind CSS 기반 컴포넌트 카탈로그',
  },
  notFound: {
    title: '페이지를 찾을 수 없습니다',
    desc: '주소가 올바른지 확인해 주세요.',
    back: '대시보드로',
  },
}

export default ko
