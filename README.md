# 🏓 PingPong Match Logger

โปรแกรมจำลองการแข่งขันปิงปองระหว่างผู้เล่น A และ B ด้วยการสื่อสารระหว่าง Goroutine พร้อมบันทึกข้อมูลการแข่งขันลงใน MongoDB, Redis และไฟล์ CSV

## 🔧 ความสามารถหลัก

- จำลองการแข่งปิงปองโดยใช้ Goroutines (`playerA`, `playerB`)
- เก็บข้อมูลการแข่งขันลงใน:
  - MongoDB
  - Redis
  - CSV (`match_log.csv`)
- มี API แสดงสถานะของระบบ และสถิติต่าง ๆ
- รองรับการดูข้อมูลผ่าน HTTP Server (Port 8888 และ 8889)



## 🚀 วิธีเริ่มต้นใช้งาน

### 1. โคลนโปรเจกต์จาก Git

```bash
git clone https://github.com/kunaaa123/pingpong.git
cd pingpong
Go run pingpong.go