import { BrowserRouter, Routes, Route } from "react-router-dom";
import AppLayout from "@/shared/components/AppLayout";
import ProtectedRoute from "@/shared/components/ProtectedRoute";
import LoginPage from "@/pages/LoginPage";
import RegisterPage from "@/pages/RegisterPage";
import ForgotPasswordPage from "@/pages/ForgotPasswordPage";
import HomePage from "@/pages/HomePage";
import ProfilePage from "@/pages/ProfilePage";
import ChangePasswordPage from "@/pages/ChangePasswordPage";
import MissingListPage from "@/pages/MissingListPage";
import MissingFormPage from "@/pages/MissingFormPage";
import DashboardPage from "@/pages/DashboardPage";
import HomelessListPage from "@/pages/HomelessListPage";
import HomelessFormPage from "@/pages/HomelessFormPage";
import FaqPage from "@/pages/FaqPage";
import TermsPage from "@/pages/TermsPage";
import PrivacyPage from "@/pages/PrivacyPage";
import NotFoundPage from "@/pages/NotFoundPage";

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/login" element={<LoginPage />} />
        <Route path="/register" element={<RegisterPage />} />
        <Route path="/forgot-password" element={<ForgotPasswordPage />} />

        <Route
          element={
            <ProtectedRoute>
              <AppLayout />
            </ProtectedRoute>
          }
        >
          <Route path="/" element={<HomePage />} />
          <Route path="/missing" element={<MissingListPage />} />
          <Route path="/missing/new" element={<MissingFormPage />} />
          <Route path="/missing/:id/edit" element={<MissingFormPage />} />
          <Route path="/dashboard" element={<DashboardPage />} />
          <Route path="/homeless" element={<HomelessListPage />} />
          <Route path="/homeless/new" element={<HomelessFormPage />} />
          <Route path="/faq" element={<FaqPage />} />
          <Route path="/terms" element={<TermsPage />} />
          <Route path="/privacy" element={<PrivacyPage />} />
          <Route path="/profile" element={<ProfilePage />} />
          <Route path="/password" element={<ChangePasswordPage />} />
          <Route path="*" element={<NotFoundPage />} />
        </Route>
      </Routes>
    </BrowserRouter>
  );
}
