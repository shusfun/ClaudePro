import {defineConfig} from '@playwright/test';
export default defineConfig({
 testDir:'./tests',workers:1,use:{baseURL:'http://127.0.0.1:4191',channel:process.env.PW_CHANNEL||'msedge',headless:true,viewport:{width:1160,height:820},screenshot:'only-on-failure'},
 webServer:{command:'npm run preview -- --host 127.0.0.1 --port 4191',url:'http://127.0.0.1:4191',reuseExistingServer:false},
 outputDir:'../test-results/ui',
});
