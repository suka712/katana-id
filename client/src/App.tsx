import { Route, Routes } from "react-router-dom";
import { Toaster } from "sonner";
import LandingPage from "./pages/public-pages/LandingPage";
import LoginPage from "./pages/public-pages/SignInPage";

function App() {
  return (
    <>
      <Toaster position="top-center" />
      <Routes>
        <Route path="/" element={<LandingPage />} />
        <Route path="/signin" element={<LoginPage />} />
      </Routes>
    </>
  );
}

export default App;
