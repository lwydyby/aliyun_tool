import ReactDOM from 'react-dom/client'
import App from './App.tsx'
import setRootPixel from '@arco-design/mobile-react/tools/flexible';

setRootPixel();
ReactDOM.createRoot(document.getElementById('root')!).render(<App />)
