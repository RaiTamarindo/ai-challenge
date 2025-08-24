package com.featurevoting.activities

import android.os.Bundle
import android.view.View
import androidx.appcompat.app.AppCompatActivity
import androidx.lifecycle.lifecycleScope
import com.featurevoting.api.ApiClient
import com.featurevoting.databinding.ActivityCreateFeatureBinding
import com.featurevoting.models.CreateFeatureRequest
import com.featurevoting.models.ErrorResponse
import com.featurevoting.utils.hideKeyboard
import com.featurevoting.utils.showToast
import com.google.gson.Gson
import kotlinx.coroutines.launch

class CreateFeatureActivity : AppCompatActivity() {
    
    private lateinit var binding: ActivityCreateFeatureBinding
    
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        binding = ActivityCreateFeatureBinding.inflate(layoutInflater)
        setContentView(binding.root)
        
        setupViews()
    }
    
    private fun setupViews() {
        setSupportActionBar(binding.toolbar)
        supportActionBar?.setDisplayHomeAsUpEnabled(true)
        
        binding.toolbar.setNavigationOnClickListener {
            onBackPressed()
        }
        
        binding.btnCancel.setOnClickListener {
            onBackPressed()
        }
        
        binding.btnCreate.setOnClickListener {
            hideKeyboard()
            createFeature()
        }
    }
    
    private fun createFeature() {
        val title = binding.etTitle.text.toString().trim()
        val description = binding.etDescription.text.toString().trim()
        
        if (!validateInput(title, description)) return
        
        showLoading(true)
        
        lifecycleScope.launch {
            try {
                val request = CreateFeatureRequest(title, description)
                val response = ApiClient.getApiService().createFeature(request)
                
                if (response.isSuccessful && response.body() != null) {
                    showToast("Feature created successfully!")
                    finish() // Go back to main activity
                } else {
                    val errorBody = response.errorBody()?.string()
                    val errorMessage = try {
                        val errorResponse = Gson().fromJson(errorBody, ErrorResponse::class.java)
                        errorResponse.error
                    } catch (e: Exception) {
                        "Failed to create feature"
                    }
                    showToast(errorMessage)
                }
            } catch (e: Exception) {
                showToast("Network error. Please check your connection.")
            } finally {
                showLoading(false)
            }
        }
    }
    
    private fun validateInput(title: String, description: String): Boolean {
        when {
            title.isEmpty() -> {
                binding.etTitle.error = "Title is required"
                binding.etTitle.requestFocus()
                return false
            }
            title.length < 3 -> {
                binding.etTitle.error = "Title must be at least 3 characters"
                binding.etTitle.requestFocus()
                return false
            }
            title.length > 100 -> {
                binding.etTitle.error = "Title must be less than 100 characters"
                binding.etTitle.requestFocus()
                return false
            }
            description.isEmpty() -> {
                binding.etDescription.error = "Description is required"
                binding.etDescription.requestFocus()
                return false
            }
            description.length < 10 -> {
                binding.etDescription.error = "Description must be at least 10 characters"
                binding.etDescription.requestFocus()
                return false
            }
            description.length > 1000 -> {
                binding.etDescription.error = "Description must be less than 1000 characters"
                binding.etDescription.requestFocus()
                return false
            }
        }
        return true
    }
    
    private fun showLoading(show: Boolean) {
        binding.progressBar.visibility = if (show) View.VISIBLE else View.GONE
        binding.btnCreate.isEnabled = !show
        binding.btnCancel.isEnabled = !show
        binding.etTitle.isEnabled = !show
        binding.etDescription.isEnabled = !show
    }
}