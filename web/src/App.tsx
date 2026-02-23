import { BrowserRouter, Routes, Route } from "react-router-dom";
import PublicLayout from "@/shared/components/PublicLayout";
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
import HeatmapPage from "@/pages/HeatmapPage";
import NotificationsPage from "@/pages/NotificationsPage";
import ChoicePage from "@/pages/ChoicePage";

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/login" element={<LoginPage />} />
        <Route path="/register" element={<RegisterPage />} />
        <Route path="/forgot-password" element={<ForgotPasswordPage />} />

        <Route element={<PublicLayout />}>
          {/* Public routes */}
          <Route path="/" element={<HomePage />} />
          <Route path="/missing" element={<MissingListPage />} />
          <Route path="/homeless" element={<HomelessListPage />} />
          <Route path="/heatmap" element={<HeatmapPage />} />
          <Route path="/register-choice" element={<ChoicePage />} />
          <Route path="/faq" element={<FaqPage />} />
          <Route path="/terms" element={<TermsPage />} />
          <Route path="/privacy" element={<PrivacyPage />} />

          {/* Protected routes */}
          <Route path="/homeless/new" element={<HomelessFormPage />} />
          <Route path="/missing/new" element={<ProtectedRoute><MissingFormPage /></ProtectedRoute>} />
          <Route path="/missing/:id/edit" element={<ProtectedRoute><MissingFormPage /></ProtectedRoute>} />
          <Route path="/dashboard" element={<ProtectedRoute><DashboardPage /></ProtectedRoute>} />
          <Route path="/notifications" element={<ProtectedRoute><NotificationsPage /></ProtectedRoute>} />
          <Route path="/profile" element={<ProtectedRoute><ProfilePage /></ProtectedRoute>} />
          <Route path="/password" element={<ProtectedRoute><ChangePasswordPage /></ProtectedRoute>} />

          <Route path="*" element={<NotFoundPage />} />
        </Route>
      </Routes>
    </BrowserRouter>
  );
}
