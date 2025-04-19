🏓 PingPong Match Logger
โปรแกรมจำลองการแข่งขันปิงปองระหว่างผู้เล่น A และ B โดยใช้การสื่อสารระหว่าง Goroutines พร้อมบันทึกข้อมูลการแข่งขันลงใน MongoDB, Redis, และไฟล์ CSV.

🔧 ความสามารถหลัก
จำลองการแข่งปิงปองโดยใช้ Goroutines (playerA, playerB)

เก็บข้อมูลการแข่งขันลงใน:

MongoDB

Redis

CSV (match_log.csv)

มี API สำหรับแสดงสถานะของระบบและสถิติต่าง ๆ

รองรับการดูข้อมูลผ่าน HTTP Server (Port 8888 และ 8889)

🚀 วิธีเริ่มต้นใช้งาน
1. โคลนโปรเจกต์จาก Git
bash
คัดลอก
แก้ไข
git clone https://github.com/kunaaa123/pingpong.git
2. เข้าไปในโฟลเดอร์โปรเจกต์
bash
คัดลอก
แก้ไข
cd pingpong
3. รันโปรแกรม
bash
คัดลอก
แก้ไข
go run pingpong.go
