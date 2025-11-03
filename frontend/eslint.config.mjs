import { dirname } from "path";
import { fileURLToPath } from "url";
import { FlatCompat } from "@eslint/eslintrc";

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

const compat = new FlatCompat({
  baseDirectory: __dirname,
});

const eslintConfig = [
  ...compat.extends("next/core-web-vitals", "next/typescript"),
  {
    ignores: [
      "node_modules/**",
      ".next/**",
      "out/**",
      "build/**",
      "next-env.d.ts",
    ],
  },
  {
    rules: {
      // 사용되지 않는 변수 경고 설정
      // 개발 중이거나 미래에 사용할 변수는 언더스코어 접두사로 표시
      "@typescript-eslint/no-unused-vars": [
        "warn",
        {
          argsIgnorePattern: "^_",
          varsIgnorePattern: "^_",
          caughtErrorsIgnorePattern: "^_",
        },
      ],
      // React Hook 의존성 경고는 유지 (중요한 규칙)
      "react-hooks/exhaustive-deps": "warn",
      // 접근성 관련 경고는 유지 (중요한 규칙)
      "jsx-a11y/role-supports-aria-props": "warn",
    },
  },
];

export default eslintConfig;
