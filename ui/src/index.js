import React from 'react';
import ReactDOM from 'react-dom/client';
import './index.css';
import App from './App';
import Register from "./Register";
import * as gateway from "./api/gateway";
import {createBrowserRouter, RouterProvider} from "react-router-dom";

const router = createBrowserRouter([
    {
        path: "/",
        element: <App />,
    },
    {
        path: "/register",
        element: <Register />,
    }
]);

const root = ReactDOM.createRoot(document.getElementById('root'));
root.render(
  <React.StrictMode>
    <RouterProvider router={router} />
  </React.StrictMode>
);

gateway.init({
    url: '/api/v1',
    getAuthorization
});

function getAuthorization(security) {
    switch(security.id) {
        case 'key': return getApiKey();
        default: console.log('default');
    }
}

function getApiKey() {
    const localUser = JSON.parse(localStorage.getItem("user"))
    if(localUser) {
        return Promise.resolve({ apiKey: localUser.token });
    } else {
        throw new Error("token not available");
    }
}