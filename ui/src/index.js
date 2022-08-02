import React from 'react';
import ReactDOM from 'react-dom/client';
import './index.css';
import App from './App';
import * as gateway from "./api/gateway";

const root = ReactDOM.createRoot(document.getElementById('root'));
root.render(
  <React.StrictMode>
    <App />
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
        console.log('getApiKey', localUser.token)
        return Promise.resolve({ apiKey: localUser.token });
    } else {
        throw new Error("token not available");
    }
}