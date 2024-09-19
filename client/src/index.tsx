import React from 'react';
import ReactDOM from 'react-dom/client';
import { createBrowserRouter, RouterProvider } from 'react-router-dom';
import Home from './pages/Home';
import Judge from './pages/judge';
import Admin from './pages/admin';
import AddProjects from './pages/admin/AddProjects';
import JudgeWelcome from './pages/judge/welcome';
import JudgeLive from './pages/judge/live';
import Project from './pages/judge/project';

import AdminSettings from './pages/admin/settings';
import Expo from './pages/Expo';

import './index.css';

const root = ReactDOM.createRoot(document.getElementById('root') as HTMLElement);

const router = createBrowserRouter([
    {
        path: '/',
        element: <Home />,
    },
    {
        path: '/judge/welcome',
        element: <JudgeWelcome />,
    },
    {
        path: '/judge/live',
        element: <JudgeLive />,
    },
    {
        path: '/judge',
        element: <Judge />,
    },
    {
        path: '/judge/project/:id',
        element: <Project />,
    },
    {
        path: '/admin',
        element: <Admin />,
    },
    {
        path: '/expo',
        element: <Expo />,
    },
    {
        path: '/admin/add-projects',
        element: <AddProjects />,
    },
    {
        path: '/admin/settings',
        element: <AdminSettings />,
    }
]);

root.render(
    <React.StrictMode>
        <RouterProvider router={router} />
    </React.StrictMode>
);
