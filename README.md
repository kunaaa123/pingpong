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

    bash
# 🏓 วิธีติดตั้งและรันโปรแกรม PingPong Match Logger

# 1. โคลนโปรเจกต์จาก Git
git clone https://github.com/kunaaa123/pingpong.git

# 2. เข้าไปในโฟลเดอร์โปรเจกต์
cd pingpong

# 3. ติดตั้ง dependencies (ถ้ามี)
go mod download

# 4. รันโปรแกรม
go run main.go

# หรือถ้าต้องการ build เป็นไฟล์ executable
go build -o pingpong
./pingpong