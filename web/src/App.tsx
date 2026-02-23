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
          <Route path="/profile" element={<ProfilePage />} />
          <Route path="/password" element={<ChangePasswordPage />} />
        </Route>
      </Routes>
    </BrowserRouter>
  );
}
