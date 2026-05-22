// frontend/src/main.js
import './style.css'; // استيراد ملف التنسيق الافتراضي (يمكن تركه فارغاً)

// 1. تعريف هيكل الواجهة (HTML String)
const appHTML = `
<div class="h-full w-full flex flex-col bg-vs-bg text-vs-text text-sm overflow-hidden">

    <!-- الشريط العلوي (Title Bar) -->
    <div class="h-8 bg-vs-sidebar flex items-center justify-between px-2 border-b border-vs-border select-none">
        <div class="flex items-center space-x-2">
            <i data-lucide="menu" class="w-4 h-4 opacity-80 hover:opacity-100 cursor-pointer"></i>
            <span class="text-xs opacity-80">File</span>
            <span class="text-xs opacity-80">Edit</span>
            <span class="text-xs opacity-80">View</span>
            <span class="text-xs opacity-80">Run</span>
            <span class="text-xs opacity-80">Terminal</span>
            <span class="text-xs opacity-80">Help</span>
        </div>
        <div class="absolute left-1/2 transform -translate-x-1/2 text-xs opacity-80">
            Wails App - Visual Studio Code
        </div>
        <div class="flex items-center space-x-2">
            <i data-lucide="minus" class="w-4 h-4 hover:bg-vs-active cursor-pointer"></i>
            <i data-lucide="square" class="w-3 h-3 hover:bg-vs-active cursor-pointer"></i>
            <i data-lucide="x" class="w-4 h-4 hover:bg-red-600 cursor-pointer"></i>
        </div>
    </div>

    <!-- المحتوى الرئيسي (Sidebar + Editor + Terminal) -->
    <div class="flex flex-1 overflow-hidden">
        
        <!-- Activity Bar (أيقونات الجانب) -->
        <div class="w-12 bg-vs-sidebar flex flex-col items-center py-2 space-y-1 border-r border-vs-border">
            <div class="p-2 border-l-2 border-vs-white bg-vs-active text-vs-white cursor-pointer">
                <i data-lucide="files" class="w-6 h-6"></i>
            </div>
            <div class="p-2 text-vs-text opacity-60 hover:opacity-100 cursor-pointer">
                <i data-lucide="search" class="w-6 h-6"></i>
            </div>
            <div class="p-2 text-vs-text opacity-60 hover:opacity-100 cursor-pointer">
                <i data-lucide="git-branch" class="w-6 h-6"></i>
            </div>
            <div class="p-2 text-vs-text opacity-60 hover:opacity-100 cursor-pointer">
                <i data-lucide="bug" class="w-6 h-6"></i>
            </div>
            <div class="mt-auto p-2 text-vs-text opacity-60 hover:opacity-100 cursor-pointer">
                <i data-lucide="settings" class="w-6 h-6"></i>
            </div>
        </div>

        <!-- Side Bar (مستكشف الملفات) -->
        <div class="w-60 bg-vs-sidebar flex flex-col border-r border-vs-border overflow-y-auto">
            <div class="p-2 text-xs uppercase tracking-wide opacity-80">Explorer</div>
            
            <div class="px-2">
                <div class="flex items-center space-x-1 py-1 hover:bg-vs-active cursor-pointer rounded-sm px-1">
                    <i data-lucide="chevron-down" class="w-4 h-4 text-vs-accent"></i>
                    <i data-lucide="folder-open" class="w-4 h-4 text-vs-accent"></i>
                    <span class="font-bold text-xs">MY-PROJECT</span>
                </div>
                
                <!-- محتوى المجلدات -->
                <div class="pl-4 space-y-0.5 text-vs-text">
                    <div class="flex items-center space-x-1 py-0.5 hover:bg-vs-active cursor-pointer rounded-sm px-1">
                        <i data-lucide="chevron-down" class="w-3 h-3"></i>
                        <i data-lucide="folder" class="w-4 h-4" style="color: #dcb67a;"></i>
                        <span>frontend</span>
                    </div>
                    <div class="pl-4">
                        <div class="flex items-center space-x-1 py-0.5 bg-vs-active rounded-sm px-1">
                            <i data-lucide="file-code" class="w-4 h-4 text-blue-400"></i>
                            <span>main.js</span>
                        </div>
                        <div class="flex items-center space-x-1 py-0.5 hover:bg-vs-active cursor-pointer rounded-sm px-1">
                            <i data-lucide="file-code" class="w-4 h-4 text-orange-400"></i>
                            <span>index.html</span>
                        </div>
                    </div>
                     <div class="flex items-center space-x-1 py-0.5 hover:bg-vs-active cursor-pointer rounded-sm px-1">
                        <i data-lucide="file-text" class="w-4 h-4 text-blue-300"></i>
                        <span>go.mod</span>
                    </div>
                </div>
            </div>
        </div>

        <!-- منطقة المحرر والترمنال -->
        <div class="flex-1 flex flex-col overflow-hidden">
            
            <!-- Tabs (التبويبات) -->
            <div class="h-9 bg-vs-bg flex items-end space-x-0 border-b border-vs-border">
                <div class="flex items-center space-x-2 bg-vs-active px-3 py-1.5 border-t-2 border-vs-accent text-vs-white text-xs cursor-pointer relative">
                    <div class="absolute left-0 top-0 bottom-0 w-1 bg-vs-accent"></div>
                    <i data-lucide="file-code" class="w-4 h-4 text-blue-400"></i>
                    <span>main.go</span>
                    <i data-lucide="x" class="w-3 h-3 opacity-50 hover:opacity-100 ml-2"></i>
                </div>
            </div>

            <!-- Editor (محرر الكود) -->
            <div class="flex-1 bg-vs-bg flex overflow-y-auto">
                <!-- أرقام الأسطر -->
                <div class="w-12 text-right pr-4 pt-2 text-vs-text opacity-50 select-none bg-vs-bg text-xs leading-6 font-mono">
                    <div>1</div><div>2</div><div>3</div><div>4</div><div>5</div><div>6</div><div>7</div><div>8</div>
                </div>
                <!-- الكود -->
                <div class="flex-1 pt-2 pl-2 text-xs leading-6 select-text font-mono">
                    <div><span class="text-vs-keyword">package</span> main</div>
                    <div></div>
                    <div><span class="text-vs-keyword">import</span> <span class="text-vs-string">"fmt"</span></div>
                    <div></div>
                    <div><span class="text-vs-keyword">func</span> <span class="text-yellow-300">main</span>() {</div>
                    <div>&nbsp;&nbsp;<span class="text-vs-variable">fmt</span>.<span class="text-yellow-200">Println</span>(<span class="text-vs-string">"Hello Wails!"</span>)</div>
                    <div>}</div>
                </div>
            </div>

            <!-- Terminal (الترمنل) -->
            <div class="h-48 bg-vs-sidebar border-t border-vs-border flex flex-col">
                <!-- Terminal Header -->
                <div class="h-9 bg-vs-active flex items-center justify-between px-4 border-b border-vs-border">
                    <div class="flex space-x-4 text-xs">
                        <span class="text-vs-white border-b-2 border-vs-accent pb-1 cursor-pointer">TERMINAL</span>
                        <span class="opacity-60 hover:opacity-100 cursor-pointer">PROBLEMS</span>
                        <span class="opacity-60 hover:opacity-100 cursor-pointer">OUTPUT</span>
                    </div>
                    <div class="flex items-center space-x-2">
                        <i data-lucide="plus" class="w-4 h-4 opacity-60 hover:opacity-100 cursor-pointer"></i>
                        <i data-lucide="x" class="w-4 h-4 opacity-60 hover:opacity-100 cursor-pointer"></i>
                    </div>
                </div>
                <!-- Terminal Body -->
                <div id="terminal-content" class="flex-1 p-2 text-xs font-mono overflow-y-auto">
                    <div class="flex">
                        <span class="text-green-400">PS C:\\Projects\\wails-app></span>
                        <span class="ml-2 text-vs-text">wails dev</span>
                    </div>
                    <div class="text-vs-text mt-1">
                        Wails v2.0.0 - A Go framework for building desktop apps<br>
                        <span class="text-green-300">INFO:</span> Serving assets from disk...<br>
                        <span class="text-green-300">INFO:</span> Application started successfully.
                    </div>
                </div>
            </div>
        </div>
    </div>

    <!-- Status Bar (الشريط السفلي) -->
    <div class="h-6 bg-vs-accent flex items-center justify-between px-2 text-xs text-white select-none">
        <div class="flex items-center space-x-3">
            <span class="flex items-center"><i data-lucide="git-branch" class="w-3 h-3 mr-1"></i> main</span>
            <span class="flex items-center opacity-80"><i data-lucide="alert-circle" class="w-3 h-3 mr-1"></i> 0 <i data-lucide="alert-triangle" class="w-3 h-3 ml-1 mr-1"></i> 0</span>
        </div>
        <div class="flex items-center space-x-3">
            <span>Ln 6, Col 2</span>
            <span>Spaces: 4</span>
            <span>UTF-8</span>
            <span>Go</span>
        </div>
    </div>
</div>
`;

// 2. حقن الهيكل داخل العنصر #app
document.getElementById('app').innerHTML = appHTML;

// 3. تشغيل الأيقونات (Lucide Icons)
// يجب استدعاء هذه الدالة بعد تحميل عناصر HTML
lucide.createIcons();

// 4. كود Wails الخاص بالتشغيل (استدعاء دالة Go كمثال)
// هذا الجزء اختياري لربط الأحداث مع الخلفية (Backend)
import * as GoFunctions from '../wailsjs/go/main/App';
// يمكنك استخدام GoFunctions هنا إذا كنت تريد ربط زر معين بدالة Go