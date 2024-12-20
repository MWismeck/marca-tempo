# marca-tempo
# Digital Point Marker (Web App) - TCC

## **Project Overview**

This project is a **Digital Point Marker Web Application** developed as part of a final course project (TCC). The application is designed to streamline the process of tracking employee work hours using a modern, efficient, and user-friendly interface. The backend is built with **Golang (Go)** to ensure performance, scalability, and ease of maintenance.

The project aims to provide an intuitive experience for employers and administrators to manage employee records, track work schedules, and ensure compliance with labor regulations.

## **Features**

- **Employee Management**: CRUD operations for managing employee records.
- **Digital Point Tracking**: Real-time recording of employee entry, break, and exit times.
- **Responsive Web App**: Accessible from both desktop and mobile devices.
- **Data Integrity & Security**: Secure handling of employee personal data.

## **API Endpoints**

The backend API provides the following routes for interacting with the employee records:

### **Employers**

### **1. List all employers**

**Route:** `GET /employers`

**Description:** Retrieves a list of all registered employees.

### **2. Create a new employer**

**Route:** `POST /employers`

**Description:** Adds a new employee to the system.

### **3. Get information about a specific employer**

**Route:** `GET /employers/:id`

**Description:** Retrieves information about a specific employee by ID.

### **4. Update an employer's information**

**Route:** `PUT /employers/:id`

**Description:** Updates the details of a specific employee by ID.

### **5. Delete an employer**

**Route:** `DELETE /employers/:id`

**Description:** Removes an employee from the system by ID.

## **Data Structure**

The **Employer** entity follows the structure below:

| **Field** | **Type** | **Description** |
| --- | --- | --- |
| `name` | `string` | Full name of the employee |
| `cpf` | `string` | Unique CPF number of the employee |
| `rg` | `string` | Identity document number (RG) |
| `email` | `string` | Employee's email address |
| `age` | `integer` | Employee's age |
| `active` | `boolean` | Whether the employee is active |
| `workload` | `integer` | Weekly work hours |

---

## **Technologies Used**

- **Frontend**: HTML, CSS, JavaScript (for user interface and interactivity)
- **Backend**: Golang (Go) (for API and business logic)
- **Database**: (To be defined) — Possible use of PostgreSQL or SQLite for lightweight storage.
- **Version Control**: Git and GitHub (for code versioning and collaboration)

## **How the Application Works**

1. **Employee Registration**: Admins can create, view, update, and delete employee records via the web app.
2. **Digital Point Tracking**: Employees register their entry, break, and exit times through an intuitive web interface.
3. **Data Management**: All actions are processed through a secure RESTful API, with the backend handling all data validation and storage.

---

## **Project Goals**

- **Efficiency**: Reduce manual processes for tracking employee work hours.
- **Usability**: Provide an easy-to-use interface for both employees and administrators.
- **Compliance**: Ensure alignment with labor regulations for accurate work hour tracking.

---

## **Future Improvements**

- **Authentication & Authorization**: Implement user roles (e.g., admin, employee) to restrict access to certain features.
- **Reports & Analytics**: Generate reports on employee attendance and work hours.
- **Notifications**: Email or SMS notifications for employees and administrators.
- **Export Data**: Allow export of employee records and work hours to CSV or PDF.