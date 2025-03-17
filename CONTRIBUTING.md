# 🚀 contributing to nubmq

Thanks for your interest in contributing to **nubmq**!  
Whether it's **fixing a bug, adding a feature, optimizing performance, or improving documentation**, every contribution is welcome.  

## 📌 How to Get Started

### **1️⃣ Reporting Issues & Feature Requests**
- If you found a bug or have a suggestion, **open an issue** [here](https://github.com/nubskr/nubmq/issues).
- Before submitting, check if a similar issue already exists.

### **2️⃣ Contributing Code**
#### 🛠 **Steps to Submit a Pull Request (PR)**:
1. **Fork the repository** and clone it locally.
2. **Create a new branch** (`git checkout -b feature-xyz`).
3. **Write your code and test it.**
4. **Push your branch** and create a PR.

#### 🏗 **Areas Needing Contributions**
✅ **Performance Enhancements** → Optimizing sharding and reducing contention.  
✅ **Custom Data Structures** → Exploring alternatives to `sync.Map` for better concurrency.  
✅ **Clustering Support** → Expanding nubmq beyond a single-node store.  

> If you're unsure what to work on, check the [open issues](https://github.com/nubskr/nubmq/issues).

---

## 📜 Code Guidelines
- Keep code **clean, simple, and idiomatic Go**.
- Avoid unnecessary dependencies.
- Write comments **only when necessary** (code should be self-explanatory).
- Use **`log.Print()` sparingly** in PRs (debug logs should be removed before merging).

---

## 📊 Running Benchmarks & Tests
If you're adding performance improvements, run the existing benchmark suite:

```sh
go test
```
