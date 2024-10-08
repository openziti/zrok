import './index.css';
import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import { createBrowserRouter, RouterProvider } from "react-router-dom";
import Overview from "./Overview.jsx";
import ShareDetail from "./ShareDetail.jsx";

const router = createBrowserRouter([
    {
        path: "/",
        element: <Overview />
    },
    {
        path: "/share/:token",
        element: <ShareDetail />
    }
]);

createRoot(document.getElementById('root')).render(
  <StrictMode>
    <RouterProvider router={router} />
  </StrictMode>,
)
