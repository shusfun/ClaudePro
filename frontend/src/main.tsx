import ReactDOM from 'react-dom/client'
import App from './App'
import {ToastProvider} from '@claudepro/ui'
import {watchSystemTheme} from './theme'

watchSystemTheme()

ReactDOM.createRoot(document.getElementById('root') as HTMLElement).render(
  <ToastProvider><App /></ToastProvider>,
)
