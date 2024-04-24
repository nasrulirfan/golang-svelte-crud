export async function load ({ fetch }) {
        const response = await fetch('http://localhost:3000/employee')
        const employees = await response.json()
        return { employees }
}