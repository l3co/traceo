import { useTranslation } from "react-i18next";
import { Globe } from "lucide-react";
import { Button } from "@/components/ui/button";

const languages = [
  { code: "pt-BR", label: "Português" },
  { code: "en", label: "English" },
] as const;

export function LanguageSwitcher() {
  const { i18n } = useTranslation();

  const current = languages.find((l) => l.code === i18n.language) ?? languages[0];
  const other = languages.find((l) => l.code !== i18n.language) ?? languages[1];

  return (
    <Button
      variant="outline"
      size="sm"
      onClick={() => i18n.changeLanguage(other.code)}
      aria-label={`Switch to ${other.label}`}
    >
      <Globe className="h-4 w-4" />
      {current.label}
    </Button>
  );
}
