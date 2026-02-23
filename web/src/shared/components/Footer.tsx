import { useTranslation } from "react-i18next";
import { Link } from "react-router-dom";

export default function Footer() {
  const { t } = useTranslation();

  return (
    <footer className="border-t bg-muted/30 py-6 mt-auto">
      <div className="container mx-auto px-4 flex flex-col sm:flex-row items-center justify-between gap-4 text-sm text-muted-foreground">
        <p>&copy; {new Date().getFullYear()} Traceo</p>
        <nav className="flex gap-4">
          <Link to="/faq" className="hover:underline">
            {t("footer.faq")}
          </Link>
          <Link to="/terms" className="hover:underline">
            {t("footer.terms")}
          </Link>
          <Link to="/privacy" className="hover:underline">
            {t("footer.privacy")}
          </Link>
        </nav>
      </div>
    </footer>
  );
}
