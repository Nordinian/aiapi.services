import React from 'react';
import ReactDOM from 'react-dom/client';
import { BrowserRouter } from 'react-router-dom';
import '@douyinfe/semi-ui/dist/css/semi.css';
import { UserProvider } from './context/User';
import 'react-toastify/dist/ReactToastify.css';
import { StatusProvider } from './context/Status';
import { ThemeProvider } from './context/Theme';
import PageLayout from './components/layout/PageLayout.js';
import './i18n/i18n.js';
import './index.css';

// 临时屏蔽Semi UI的findDOMNode警告
const originalWarn = console.warn;
console.warn = function(message, ...args) {
  if (typeof message === 'string' && message.includes('findDOMNode is deprecated')) {
    return;
  }
  originalWarn.apply(console, [message, ...args]);
};

// initialization

const root = ReactDOM.createRoot(document.getElementById('root'));
root.render(
  // <React.StrictMode>
    <StatusProvider>
      <UserProvider>
        <BrowserRouter
          future={{
            v7_startTransition: true,
            v7_relativeSplatPath: true,
          }}
        >
          <ThemeProvider>
            <PageLayout />
          </ThemeProvider>
        </BrowserRouter>
      </UserProvider>
    </StatusProvider>
  // </React.StrictMode>,
);
