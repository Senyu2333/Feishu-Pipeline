package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

var (
	// === input params start
	appID = os.Getenv("APP_ID") // app_id, required, 应用 ID
	// 应用唯一标识，创建应用后获得。有关app_id 的详细介绍。请参考通用参数https://open.feishu.cn/document/ukTMukTMukTM/uYTM5UjL2ETO14iNxkTN/terminology。
	appSecret = os.Getenv("APP_SECRET") // app_secret, required, 应用密钥
	// 应用秘钥，创建应用后获得。有关 app_secret 的详细介绍，请参考https://open.feishu.cn/document/ukTMukTMukTM/uYTM5UjL2ETO14iNxkTN/terminology。
	// === input params end
)

// 获取 tenant_access_token
func getTenantAccessToken(appID, appSecret string) (string, error) {
	url := "https://open.feishu.cn/open-apis/auth/v3/tenant_access_token/internal"
	payload := map[string]string{
		"app_id":     appID,
		"app_secret": appSecret,
	}
	payloadBytes, _ := json.Marshal(payload)

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result map[string]interface{}
	json.Unmarshal(body, &result)

	if result["code"].(float64) != 0 {
		return "", fmt.Errorf("failed to get tenant_access_token: %s", result["msg"].(string))
	}

	return result["tenant_access_token"].(string), nil
}

// 批量获取用户信息
func batchGetUsers(tenantAccessToken string, userIds []string) ([]map[string]interface{}, error) {
	url := "https://open.feishu.cn/open-apis/contact/v3/users/batch"
	q := "?user_id_type=open_id"
	for _, id := range userIds {
		q += "&user_ids=" + id
	}
	fullURL := url + q

	req, _ := http.NewRequest("GET", fullURL, nil)
	req.Header.Set("Authorization", "Bearer "+tenantAccessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("Batch get users response: %s\n", string(body))

	var result map[string]interface{}
	json.Unmarshal(body, &result)

	if result["code"].(float64) != 0 {
		return nil, fmt.Errorf("failed to batch get users: %s", result["msg"].(string))
	}

	items := result["data"].(map[string]interface{})["items"].([]interface{})
	users := make([]map[string]interface{}, len(items))
	for i, item := range items {
		users[i] = item.(map[string]interface{})
	}

	return users, nil
}

// 批量获取部门信息
func batchGetDepartments(tenantAccessToken string, departmentIds []string) ([]map[string]interface{}, error) {
	url := "https://open.feishu.cn/open-apis/contact/v3/departments/batch"
	q := "?department_id_type=open_department_id&user_id_type=open_id"
	for _, id := range departmentIds {
		q += "&department_ids=" + id
	}
	fullURL := url + q

	req, _ := http.NewRequest("GET", fullURL, nil)
	req.Header.Set("Authorization", "Bearer "+tenantAccessToken)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("Batch get departments response: %s\n", string(body))

	var result map[string]interface{}
	json.Unmarshal(body, &result)

	if result["code"].(float64) != 0 {
		return nil, fmt.Errorf("failed to batch get departments: %s", result["msg"].(string))
	}

	items := result["data"].(map[string]interface{})["items"].([]interface{})
	departments := make([]map[string]interface{}, len(items))
	for i, item := range items {
		departments[i] = item.(map[string]interface{})
	}

	return departments, nil
}

// 发送消息给用户
func sendMessage(tenantAccessToken, receiveID, msgType, content string) (map[string]interface{}, error) {
	url := "https://open.feishu.cn/open-apis/im/v1/messages?receive_id_type=open_id"

	message := map[string]string{
		"receive_id": receiveID,
		"msg_type":   msgType,
		"content":    content,
	}
	payloadBytes, _ := json.Marshal(message)

	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(payloadBytes))
	req.Header.Set("Authorization", "Bearer "+tenantAccessToken)
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Printf("Send message response: %s\n", string(body))

	var result map[string]interface{}
	json.Unmarshal(body, &result)

	if result["code"].(float64) != 0 {
		return nil, fmt.Errorf("failed to send message: %s", result["msg"].(string))
	}

	return result["data"].(map[string]interface{}), nil
}

func main() {
	// 获取 tenant_access_token
	tenantAccessToken, err := getTenantAccessToken(appID, appSecret)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err.Error())
		return
	}
	fmt.Printf("Successfully obtained tenant_access_token\n")

	// 示例：假设我们有一组产品部门的用户ID
	productDepartmentUserIDs := []string{
		"ou_1234567890abcdef1234567890abcdef", // 示例产品经理ID
		"ou_0987654321fedcba0987654321fedcba", // 示例前端工程师ID
		"ou_abcdef1234567890abcdef1234567890", // 示例后端工程师ID
	}

	// 步骤1: 批量获取用户基本信息
	users, err := batchGetUsers(tenantAccessToken, productDepartmentUserIDs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ERROR: %s\n", err.Error())
		return
	}
	fmt.Printf("Successfully retrieved %d users\n", len(users))

	// 步骤2: 对每个用户获取详细信息并确定其身份
	for _, user := range users {
		openID := user["open_id"].(string)
		jobTitle := ""
		if jt, ok := user["job_title"]; ok && jt != nil {
			jobTitle = jt.(string)
		}
		departmentIDs := user["department_ids"].([]interface{})

		fmt.Printf("User: %s, Job Title: %s, Department Count: %d\n", openID, jobTitle, len(departmentIDs))

		// 步骤3: 获取用户所属部门的详细信息
		var deptIDs []string
		for _, deptID := range departmentIDs {
			deptIDs = append(deptIDs, deptID.(string))
		}

		if len(deptIDs) > 0 {
			departments, err := batchGetDepartments(tenantAccessToken, deptIDs)
			if err != nil {
				fmt.Fprintf(os.Stderr, "ERROR getting departments for user %s: %s\n", openID, err.Error())
				continue
			}

			for _, dept := range departments {
				deptName := dept["name"].(string)
				deptID := dept["open_department_id"].(string)
				fmt.Printf("  Department: %s (ID: %s)\n", deptName, deptID)
			}
		}

		// 步骤4: 根据职位确定用户身份（前端、后端、产品经理等）
		// 注意：这里需要根据实际业务需求定义职位与角色的映射关系
		// 参考文档中未提供此信息
		role := "unknown"
		switch {
		case contains([]string{"产品经理", "产品总监", "Product Manager"}, jobTitle):
			role = "product_manager"
		case contains([]string{"前端", "前端开发", "Frontend"}, jobTitle):
			role = "frontend_developer"
		case contains([]string{"后端", "后端开发", "Backend"}, jobTitle):
			role = "backend_developer"
		}

		fmt.Printf("  Determined role: %s\n", role)

		// 步骤5: 在实际应用中，这里会调用大模型拆解需求
		// 然后根据拆解结果和用户角色，向相应用户推送任务
		// 以下为示例消息推送

		if role != "unknown" {
			// 示例：向用户发送任务分配消息
			content := fmt.Sprintf(`{"text":"您被分配了一个新任务，请查看相关需求文档。角色：%s"}`, role)
			_, err := sendMessage(tenantAccessToken, openID, "text", content)
			if err != nil {
				fmt.Fprintf(os.Stderr, "ERROR sending message to user %s: %s\n", openID, err.Error())
			} else {
				fmt.Printf("Successfully sent task message to user %s\n", openID)
			}
		}
	}
}

// 辅助函数：检查字符串是否在数组中
func contains(arr []string, str string) bool {
	for _, a := range arr {
		if a == str {
			return true
		}
	}
	return false
}
