package com.featurevoting.activities

import android.content.Intent
import android.os.Bundle
import android.view.View
import androidx.appcompat.app.AppCompatActivity
import androidx.lifecycle.lifecycleScope
import com.featurevoting.R
import com.featurevoting.api.ApiClient
import com.featurevoting.databinding.ActivityLoginBinding
import com.featurevoting.models.ErrorResponse
import com.featurevoting.models.LoginRequest
import com.featurevoting.utils.PreferenceManager
import com.featurevoting.utils.hideKeyboard
import com.featurevoting.utils.isValidEmail
import com.featurevoting.utils.showToast
import com.google.gson.Gson
import kotlinx.coroutines.launch

class LoginActivity : AppCompatActivity() {
    
    private lateinit var binding: ActivityLoginBinding
    private lateinit var preferenceManager: PreferenceManager
    
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        binding = ActivityLoginBinding.inflate(layoutInflater)
        setContentView(binding.root)
        
        preferenceManager = PreferenceManager(this)
        ApiClient.initialize(preferenceManager)
        
        // Check if already logged in
        if (preferenceManager.isLoggedIn()) {
            startMainActivity()
            return
        }
        
        setupViews()
    }
    
    private fun setupViews() {
        binding.btnLogin.setOnClickListener {
            hideKeyboard()
            performLogin()
        }
        
        // Pre-fill for testing (remove in production)
        binding.etEmail.setText("admin@example.com")
        binding.etPassword.setText("admin123")
    }
    
    private fun performLogin() {
        val email = binding.etEmail.text.toString().trim()
        val password = binding.etPassword.text.toString().trim()
        
        if (!validateInput(email, password)) return
        
        showLoading(true)
        
        lifecycleScope.launch {
            try {
                val loginRequest = LoginRequest(email, password)
                val response = ApiClient.getApiService().login(loginRequest)
                
                if (response.isSuccessful && response.body() != null) {
                    val loginResponse = response.body()!!
                    
                    // Save user data
                    preferenceManager.saveAuthToken(loginResponse.token)
                    preferenceManager.saveUserInfo(
                        loginResponse.user.id,
                        loginResponse.user.username,
                        loginResponse.user.email
                    )
                    
                    showToast("Welcome, ${loginResponse.user.username}!")
                    startMainActivity()
                } else {
                    val errorBody = response.errorBody()?.string()
                    val errorMessage = try {
                        val errorResponse = Gson().fromJson(errorBody, ErrorResponse::class.java)
                        errorResponse.error
                    } catch (e: Exception) {
                        "Login failed"
                    }
                    showError(errorMessage)
                }
            } catch (e: Exception) {
                showError("Network error. Please check your connection.")
            } finally {
                showLoading(false)
            }
        }
    }
    
    private fun validateInput(email: String, password: String): Boolean {
        when {
            email.isEmpty() -> {
                binding.etEmail.error = "Email is required"
                return false
            }
            !email.isValidEmail() -> {
                binding.etEmail.error = "Invalid email format"
                return false
            }
            password.isEmpty() -> {
                binding.etPassword.error = "Password is required"
                return false
            }
            password.length < 6 -> {
                binding.etPassword.error = "Password must be at least 6 characters"
                return false
            }
        }
        return true
    }
    
    private fun showLoading(show: Boolean) {
        binding.progressBar.visibility = if (show) View.VISIBLE else View.GONE
        binding.btnLogin.isEnabled = !show
        binding.etEmail.isEnabled = !show
        binding.etPassword.isEnabled = !show
        
        if (show) {
            binding.tvError.visibility = View.GONE
        }
    }
    
    private fun showError(message: String) {
        binding.tvError.text = message
        binding.tvError.visibility = View.VISIBLE
    }
    
    private fun startMainActivity() {
        startActivity(Intent(this, MainActivity::class.java))
        finish()
    }
}