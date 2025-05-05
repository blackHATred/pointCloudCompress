import * as THREE from 'three';
import { OrbitControls } from 'three/examples/jsm/controls/OrbitControls';

const scene = new THREE.Scene();
const camera = new THREE.PerspectiveCamera(75, window.innerWidth/window.innerHeight, 0.1, 1000);
camera.position.z = 20;

const renderer = new THREE.WebGLRenderer();
renderer.setSize(window.innerWidth, window.innerHeight);
document.body.appendChild(renderer.domElement);

// Добавляем элементы управления камерой
const controls = new OrbitControls(camera, renderer.domElement);
controls.enableDamping = true; // Плавность перемещения
controls.dampingFactor = 0.05;
controls.screenSpacePanning = false;
controls.minDistance = 1;
controls.maxDistance = 500;

let geometry = new THREE.BufferGeometry();
// let material = new THREE.PointsMaterial({ size: 0.1, color: 0xffffff });
let material = new THREE.PointsMaterial({
    size: 0.1,
    vertexColors: true, // Включаем цвет для каждой точки
    transparent: true,  // Поддержка прозрачности
    opacity: 0.8        // Базовая прозрачность
});
let points = new THREE.Points(geometry, material);
scene.add(points);

function updatePoints(buffer) {
    const f32 = new Float32Array(buffer);
    const vertexCount = f32.length / 4; // XYZI формат (4 значения на точку)

    const positions = new Float32Array(vertexCount * 3);
    const colors = new Float32Array(vertexCount * 3);

    // Определяем максимальные значения для нормализации
    let maxDist = 0;
    let maxIntensity = 0;

    for (let i = 0; i < f32.length; i += 4) {
        const x = f32[i];
        const y = f32[i + 1];
        const z = f32[i + 2];
        const intensity = f32[i + 3];

        const dist = Math.sqrt(x*x + y*y + z*z);
        maxDist = Math.max(maxDist, dist);
        maxIntensity = Math.max(maxIntensity, intensity);
    }

    // Заполняем массивы позиций и цветов
    for (let i = 0, j = 0, c = 0; i < f32.length; i += 4) {
        const x = f32[i];
        const y = f32[i + 1];
        const z = f32[i + 2];
        const intensity = f32[i + 3];

        positions[j++] = x;
        positions[j++] = y;
        positions[j++] = z;

        // Нормализуем значения
        const dist = Math.sqrt(x*x + y*y + z*z);
        const normDist = maxDist > 0 ? dist / maxDist : 0;
        const normIntensity = maxIntensity > 0 ? intensity / maxIntensity : 1.0;

        // Цвет зависит от расстояния (красный → синий)
        // Яркость зависит от интенсивности
        colors[c++] = normDist * normIntensity;         // R
        colors[c++] = 0.2 * normIntensity;              // G
        colors[c++] = (1 - normDist) * normIntensity;   // B
    }

    geometry.setAttribute('position', new THREE.BufferAttribute(positions, 3));
    geometry.setAttribute('color', new THREE.BufferAttribute(colors, 3));

    geometry.computeBoundingSphere();
}

const socket = new WebSocket('ws://localhost:8080/ws');
socket.binaryType = 'arraybuffer';
socket.onmessage = (event) => updatePoints(event.data);

// Добавляем обработку изменения размера окна
window.addEventListener('resize', () => {
    camera.aspect = window.innerWidth / window.innerHeight;
    camera.updateProjectionMatrix();
    renderer.setSize(window.innerWidth, window.innerHeight);
});

function animate() {
    requestAnimationFrame(animate);
    controls.update(); // Обновляем элементы управления
    renderer.render(scene, camera);
}
animate();