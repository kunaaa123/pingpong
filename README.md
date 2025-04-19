# 🏓 PingPong Match Logger

โปรแกรมจำลองการแข่งขันปิงปองระหว่างผู้เล่น A และ B โดยใช้ Goroutine พร้อมบันทึกข้อมูลการแข่งขันลงใน MongoDB, Redis และไฟล์ CSV

---

## 🔧 ความสามารถหลัก

- จำลองการแข่งขันระหว่าง `playerA` และ `playerB` ด้วย Goroutines  
- บันทึกผลการแข่งขันลงใน:
  - 🗃️ MongoDB
  - 🚀 Redis
  - 📝 ไฟล์ CSV (`match_log.csv`)
- มี API แสดงสถานะของระบบและสถิติต่าง ๆ
- มี HTTP Server ให้ดูข้อมูลได้ที่พอร์ต:
  - 🌐 `8888`
  - 🌐 `8889`

---

## 🚀 วิธีเริ่มต้นใช้งาน

### 1️⃣ โคลนโปรเจกต์จาก Git

```bash
git clone https://github.com/kunaaa123/pingpong.git
2️⃣ เข้าไปในโฟลเดอร์โปรเจกต์
bash
คัดลอก
แก้ไข
cd pingpong
3️⃣ รันโปรแกรม
bash
คัดลอก
แก้ไข
go run pingpong.go