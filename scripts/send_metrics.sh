#!/bin/bash

# Đọc thông tin từ file .env nếu tồn tại
if [ -f ".env" ]; then
    # Đọc file .env và export các biến, xử lý dấu ngoặc kép
    while IFS='=' read -r key value; do
        # Bỏ qua dòng trống và comment
        [[ -z "$key" || "$key" =~ ^[[:space:]]*# ]] && continue
        # Loại bỏ khoảng trắng thừa
        key=$(echo "$key" | xargs)
        value=$(echo "$value" | xargs)
        # Export biến
        export "$key"="$value"
    done < .env
else
    echo "File .env không tồn tại. Vui lòng chạy register.sh trước."
    exit 1
fi

# Kiểm tra các biến cần thiết
if [ -z "$SERVER_ID" ] || [ -z "$INTERVAL_TIME" ]; then
    echo "Thiếu thông tin SERVER_ID hoặc INTERVAL_TIME trong file .env"
    exit 1
fi

# Đọc access token
if [ -f "/data/access_token.txt" ]; then
    ACCESS_TOKEN=$(cat /data/access_token.txt)
else
    echo "Không tìm thấy access token. Vui lòng chạy register.sh trước."
    exit 1
fi

# Hàm lấy thông tin CPU (%)
get_cpu_usage() {
    top -bn1 | grep "Cpu(s)" | awk '{print $2}' | awk -F'%' '{print $1}' | sed 's/[^0-9.]//g' | awk '{print int($1)}'
}

# Hàm lấy thông tin RAM (%)
get_ram_usage() {
    free | grep Mem | awk '{printf "%.0f", $3/$2 * 100.0}'
}

# Hàm lấy thông tin Disk (%)
get_disk_usage() {
    df / | tail -1 | awk '{print $5}' | sed 's/%//'
}

# Hàm gửi metrics
send_metrics() {
    local cpu_usage=$(get_cpu_usage)
    local ram_usage=$(get_ram_usage)
    local disk_usage=$(get_disk_usage)
    local timestamp=$(date -u +"%Y-%m-%dT%H:%M:%S.000Z")
    
    echo "Gửi metrics - CPU: ${cpu_usage}%, RAM: ${ram_usage}%, Disk: ${disk_usage}%"

    RESPONSE=$(curl -s -X POST http://$HOST_IP:8080/api/v1/servers/monitoring \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer $ACCESS_TOKEN" \
        -d "{
            \"server_id\": \"$SERVER_ID\",
            \"cpu\": $cpu_usage,
            \"ram\": $ram_usage,
            \"disk\": $disk_usage,
            \"timestamp\": \"$timestamp\"
        }")

    echo "Response: $RESPONSE"
    
    # Kiểm tra nếu token hết hạn (status 401)
    if echo "$RESPONSE" | grep -q "401\|unauthorized\|token"; then
        echo "Token có thể đã hết hạn. Thử refresh token..."
        refresh_token
        return $?
    fi
    
    return 0
}

# Hàm refresh token
refresh_token() {
    if [ -f "/data/refresh_token.txt" ]; then
        REFRESH_TOKEN=$(cat /data/refresh_token.txt)
        
        RESPONSE=$(curl -s -X POST http://$HOST_IP:8080/api/v1/auth/refresh \
            -H "Content-Type: application/json" \
            -d "{
                \"refresh_token\": \"$REFRESH_TOKEN\"
            }")
        
        NEW_ACCESS_TOKEN=$(echo "$RESPONSE" | jq -r '.data.access_token')
        if [ "$NEW_ACCESS_TOKEN" != "null" ] && [ "$NEW_ACCESS_TOKEN" != "" ]; then
            echo "$NEW_ACCESS_TOKEN" > /data/access_token.txt
            ACCESS_TOKEN="$NEW_ACCESS_TOKEN"
            echo "Token đã được refresh thành công"
            return 0
        else
            echo "Không thể refresh token. Response: $RESPONSE"
            return 1
        fi
    else
        echo "Không tìm thấy refresh token"
        return 1
    fi
}

# Hàm chính
main() {
    echo "Bắt đầu gửi metrics cho server $SERVER_ID với interval ${INTERVAL_TIME}s"
    
    # Kiểm tra jq có sẵn không
    if ! command -v jq &> /dev/null; then
        echo "jq chưa được cài đặt. Vui lòng cài đặt jq để parse JSON."
        exit 1
    fi
    
    # Gửi metrics đầu tiên ngay lập tức
    send_metrics
    
    # Lặp vô hạn với interval time
    while true; do
        sleep $INTERVAL_TIME
        send_metrics
        
        # Nếu gửi thất bại nhiều lần, có thể cần dừng
        if [ $? -ne 0 ]; then
            echo "Gặp lỗi khi gửi metrics. Thử lại sau ${INTERVAL_TIME}s..."
        fi
    done
}

# Xử lý tín hiệu để dừng script một cách graceful
trap 'echo "Dừng gửi metrics..."; exit 0' SIGINT SIGTERM

# Chạy hàm chính
main "$@"